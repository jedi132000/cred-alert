package queue_test

import (
	"errors"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/pivotal-golang/lager/lagertest"

	"cred-alert/github/githubfakes"
	"cred-alert/models"
	"cred-alert/models/modelsfakes"
	"cred-alert/queue"
	"cred-alert/queue/queuefakes"
)

var _ = Describe("Ancestry Scan Job", func() {
	var (
		logger *lagertest.TestLogger

		taskQueue        *queuefakes.FakeQueue
		client           *githubfakes.FakeClient
		commitRepository *modelsfakes.FakeCommitRepository

		plan queue.AncestryScanPlan
		job  *queue.AncestryScanJob
	)

	BeforeEach(func() {
		plan = queue.AncestryScanPlan{
			Owner:      "owner",
			Repository: "repo",
			SHA:        "sha",
		}

		taskQueue = &queuefakes.FakeQueue{}
		client = &githubfakes.FakeClient{}
		commitRepository = &modelsfakes.FakeCommitRepository{}
		logger = lagertest.NewTestLogger("ancestry-scan")
	})

	JustBeforeEach(func() {
		job = queue.NewAncestryScanJob(plan, commitRepository, client, taskQueue)
	})

	var ItMarksTheCommitAsSeen = func() {
		It("marks the commit as seen", func() {
			err := job.Run(logger)
			Expect(err).NotTo(HaveOccurred())

			Expect(commitRepository.RegisterCommitCallCount()).To(Equal(1))
			_, registeredCommit := commitRepository.RegisterCommitArgsForCall(0)
			Expect(registeredCommit).To(Equal(&models.Commit{
				Owner:      "owner",
				Repository: "repo",
				SHA:        "sha",
				// TODO: timestamp
			}))
		})
	}

	var ItStopsAndDoesNotEnqueueAnyMoreWork = func() {
		It("stops and does not enqueue any more work", func() {
			err := job.Run(logger)
			Expect(err).NotTo(HaveOccurred())

			Expect(taskQueue.EnqueueCallCount()).To(BeZero())
			Expect(commitRepository.RegisterCommitCallCount()).To(BeZero())
		})
	}

	var ItReturnsAndLogsAnError = func(expectedError error) {
		It("returns and logs an error", func() {
			err := job.Run(logger)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedError))
			Expect(logger).To(gbytes.Say("scanning-ancestry.failed"))
		})
	}

	var ItDoesNotRegisterCommit = func() {
		It("returns and logs an error", func() {
			job.Run(logger)
			Expect(commitRepository.RegisterCommitCallCount()).To(BeZero())
		})
	}

	Describe("running the job", func() {
		Context("when the commit repository has an error finding a commit", func() {
			expectedError := errors.New("client repository error")

			BeforeEach(func() {
				commitRepository.IsCommitRegisteredReturns(false, expectedError)
			})

			ItReturnsAndLogsAnError(expectedError)
			ItDoesNotRegisterCommit()
		})

		Context("when we have not previously scanned the commit", func() {
			BeforeEach(func() {
				commitRepository.IsCommitRegisteredReturns(false, nil)
			})

			Context("when we have not reached the maximum scan depth", func() {
				BeforeEach(func() {
					plan.Depth = 5
				})

				Context("when the github client returns an error", func() {
					expectedError := errors.New("client error")

					BeforeEach(func() {
						client.ParentsReturns(nil, expectedError)
					})

					ItReturnsAndLogsAnError(expectedError)
					ItDoesNotRegisterCommit()
				})

				Context("when the commit has parents", func() {
					BeforeEach(func() {
						client.ParentsReturns([]string{
							"abcdef",
							"123456",
							"789aee",
						}, nil)
					})

					Context("when the task queue returns an error enqueueing diffs", func() {
						expectedError := errors.New("queue error")

						BeforeEach(func() {
							taskQueue.EnqueueStub = func(task queue.Task) error {
								if task.Type() == "diff-scan" {
									return expectedError
								}
								return nil
							}
						})

						ItReturnsAndLogsAnError(expectedError)
						ItDoesNotRegisterCommit()
					})

					It("scans the diffs between the current commit and its parents", func() {
						err := job.Run(logger)
						Expect(err).NotTo(HaveOccurred())

						Expect(taskQueue.EnqueueCallCount()).To(Equal(6))

						task := taskQueue.EnqueueArgsForCall(0)
						Expect(task.Type()).To(Equal("diff-scan"))
						Expect(task.Payload()).To(MatchJSON(`
							{
								"owner": "owner",
								"repository": "repo",
								"from": "abcdef",
								"to": "sha"
							}
						`))

						task = taskQueue.EnqueueArgsForCall(2)
						Expect(task.Type()).To(Equal("diff-scan"))
						Expect(task.Payload()).To(MatchJSON(`
							{
								"owner": "owner",
								"repository": "repo",
								"from": "123456",
								"to": "sha"
							}
						`))

						task = taskQueue.EnqueueArgsForCall(4)
						Expect(task.Type()).To(Equal("diff-scan"))
						Expect(task.Payload()).To(MatchJSON(`
							{
								"owner": "owner",
								"repository": "repo",
								"from": "789aee",
								"to": "sha"
							}
						`))
					})

					Context("when the task queue returns an error enqueueing ancestry scans", func() {
						expectedError := errors.New("disaster")
						BeforeEach(func() {
							taskQueue.EnqueueStub = func(task queue.Task) error {
								if task.Type() == "ancestry-scan" {
									return expectedError
								}
								return nil
							}
						})

						ItReturnsAndLogsAnError(expectedError)
						ItDoesNotRegisterCommit()
					})

					Context("when the commit repository returns an error registering the commit", func() {
						expectedError := errors.New("disaster")
						BeforeEach(func() {
							commitRepository.RegisterCommitReturns(expectedError)
						})

						ItReturnsAndLogsAnError(expectedError)
					})

					It("queues an ancestry-scan for each parent commit with one less depth", func() {
						err := job.Run(logger)
						Expect(err).NotTo(HaveOccurred())

						Expect(taskQueue.EnqueueCallCount()).To(Equal(6))

						task := taskQueue.EnqueueArgsForCall(1)
						Expect(task.Type()).To(Equal("ancestry-scan"))
						Expect(task.Payload()).To(MatchJSON(`
							{
								"owner": "owner",
								"repository": "repo",
								"commit-timestamp": 0,
								"sha": "abcdef",
								"depth": 4
							}
						`))

						task = taskQueue.EnqueueArgsForCall(3)
						Expect(task.Type()).To(Equal("ancestry-scan"))
						Expect(task.Payload()).To(MatchJSON(`
							{
								"owner": "owner",
								"repository": "repo",
								"commit-timestamp": 0,
								"sha": "123456",
								"depth": 4
							}
						`))

						task = taskQueue.EnqueueArgsForCall(5)
						Expect(task.Type()).To(Equal("ancestry-scan"))
						Expect(task.Payload()).To(MatchJSON(`
							{
								"owner": "owner",
								"repository": "repo",
								"commit-timestamp": 0,
								"sha": "789aee",
								"depth": 4
							}
						`))
					})

					ItMarksTheCommitAsSeen()
				})

				// Does initialCommit.parents() return 0000000? or empty list
				XContext("when the current commit is the initial commit", func() {
					BeforeEach(func() {
						plan.SHA = strings.Repeat("0", 40)
					})

					ItStopsAndDoesNotEnqueueAnyMoreWork()
				})
			})

			Context("when we have reached the maximum scan depth", func() {
				BeforeEach(func() {
					plan.Depth = 0
					// Fail if it tries to enqueue more tasks
					taskQueue.EnqueueStub = func(task queue.Task) error {
						Expect(task.Type()).To(Equal("ref-scan"))
						return nil
					}
				})

				It("enqueues a ref scan of the commit", func() {
					err := job.Run(logger)
					Expect(err).NotTo(HaveOccurred())

					Expect(taskQueue.EnqueueCallCount()).To(Equal(1))

					task := taskQueue.EnqueueArgsForCall(0)
					Expect(task.Type()).To(Equal("ref-scan"))
					Expect(task.Payload()).To(MatchJSON(`
							{
								"owner": "owner",
								"repository": "repo",
								"ref": "sha"
							}
						`))
				})

				ItMarksTheCommitAsSeen()

				XIt("sends a notification saying that it ran out of depth", func() {
				})

				It("does not look for any more parents", func() {
					Expect(client.ParentsCallCount()).To(Equal(0))
				})

				Context("When there is an error registering a commit", func() {
					expectedError := errors.New("disaster")
					BeforeEach(func() {
						commitRepository.RegisterCommitReturns(expectedError)
					})

					ItReturnsAndLogsAnError(expectedError)
				})

				Context("when there is an error enqueuing a ref scan", func() {
					expectedError := errors.New("disaster")

					BeforeEach(func() {
						taskQueue.EnqueueStub = func(task queue.Task) error {
							Expect(task.Type()).To(Equal("ref-scan"))
							return expectedError
						}
					})

					ItReturnsAndLogsAnError(expectedError)
					ItDoesNotRegisterCommit()
				})
			})
		})

		Context("when we have previously scanned the commit", func() {
			BeforeEach(func() {
				commitRepository.IsCommitRegisteredReturns(true, nil)
			})

			ItStopsAndDoesNotEnqueueAnyMoreWork()
		})

		Context("when there is an error checking if we have scanned the commit", func() {
			BeforeEach(func() {
				commitRepository.IsCommitRegisteredReturns(false, errors.New("disaster"))
			})

			It("stops and does not enqueue any more work", func() {
				err := job.Run(logger)
				Expect(err).To(MatchError("disaster"))

				Expect(taskQueue.EnqueueCallCount()).To(BeZero())
				Expect(commitRepository.RegisterCommitCallCount()).To(BeZero())
			})
		})
	})
})