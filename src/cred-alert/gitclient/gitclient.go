package gitclient

import (
	"bufio"
	"bytes"
	"cred-alert/mimetype"
	"cred-alert/scanners"
	"cred-alert/scanners/filescanner"
	"cred-alert/sniff"
	"errors"
	"fmt"
	"strings"

	"code.cloudfoundry.org/lager"

	git "gopkg.in/libgit2/git2go.v24"
)

const defaultRemoteName = "origin"

var ErrInterrupted = errors.New("interrupted")

type client struct {
	privateKeyPath string
	publicKeyPath string
}

//go:generate counterfeiter . Client

type Client interface {
	BranchTargets(string) (map[string]string, error)
	Clone(string, string) (*git.Repository, error)
	GetParents(*git.Repository, *git.Oid) ([]*git.Oid, error)
	Fetch(string) (map[string][]*git.Oid, error)
	HardReset(string, *git.Oid) error
	Diff(repositoryPath string, a, b *git.Oid) (string, error)
	BranchCredentialCounts(lager.Logger, string, sniff.Sniffer, git.BranchType) (map[string]uint, error)
}

func New(privateKeyPath, publicKeyPath string) *client {
	return &client{
		privateKeyPath: privateKeyPath,
		publicKeyPath: publicKeyPath,
	}
}

func (c *client) BranchTargets(repositoryPath string) (map[string]string, error) {
	repo, err := git.OpenRepository(repositoryPath)
	if err != nil {
		return nil, err
	}
	defer repo.Free()

	it, err := repo.NewBranchIterator(git.BranchAll)
	if err != nil {
		return nil, err
	}

	var branch *git.Branch
	branches := map[string]string{}
	for {
		branch, _, err = it.Next()
		if err != nil {
			break
		}

		branchName, err := branch.Name()
		if err != nil {
			break
		}

		target := branch.Target()
		if target == nil { // origin/HEAD has no target
			continue
		}

		branches[branchName] = branch.Target().String()
	}

	if branch != nil {
		branch.Free()
	}

	return branches, nil
}

func (c *client) Clone(sshURL, dest string) (*git.Repository, error) {
	cloneOptions := &git.CloneOptions{
		FetchOptions: newFetchOptions(c.privateKeyPath, c.publicKeyPath),
	}

	return git.Clone(sshURL, dest, cloneOptions)
}

func (c *client) GetParents(repo *git.Repository, child *git.Oid) ([]*git.Oid, error) {
	object, err := repo.Lookup(child)
	if err != nil {
		return nil, err
	}
	defer object.Free()

	commit, err := object.AsCommit()
	if err != nil {
		return nil, err
	}
	defer commit.Free()

	var parents []*git.Oid
	var i uint
	for i = 0; i < commit.ParentCount(); i++ {
		parents = append(parents, commit.ParentId(i))
	}

	return parents, nil
}

func (c *client) Fetch(repositoryPath string) (map[string][]*git.Oid, error) {
	repo, err := git.OpenRepository(repositoryPath)
	if err != nil {
		return nil, err
	}
	defer repo.Free()

	remote, err := repo.Remotes.Lookup(defaultRemoteName)
	if err != nil {
		return nil, err
	}
	defer remote.Free()

	changes := map[string][]*git.Oid{}
	updateTipsCallback := func(refname string, a *git.Oid, b *git.Oid) git.ErrorCode {
		changes[refname] = []*git.Oid{a.Copy(), b.Copy()}
		return 0
	}

	fetchOptions := newFetchOptions(c.privateKeyPath, c.publicKeyPath)
    fetchOptions.RemoteCallbacks.UpdateTipsCallback = updateTipsCallback

	var msg string
	err = remote.Fetch([]string{}, fetchOptions, msg)
	if err != nil {
		return nil, err
	}

	return changes, nil
}

func (c *client) HardReset(repositoryPath string, oid *git.Oid) error {
	repo, err := git.OpenRepository(repositoryPath)
	if err != nil {
		return err
	}
	defer repo.Free()

	object, err := repo.Lookup(oid)
	if err != nil {
		return err
	}
	defer object.Free()

	commit, err := object.AsCommit()
	if err != nil {
		return err
	}
	defer commit.Free()

	return repo.ResetToCommit(commit, git.ResetHard, &git.CheckoutOpts{
		Strategy: git.CheckoutForce,
	})
}

func (c *client) Diff(repositoryPath string, parent, child *git.Oid) (string, error) {
	repo, err := git.OpenRepository(repositoryPath)
	if err != nil {
		return "", err
	}
	defer repo.Free()

	var aTree *git.Tree
	if parent != nil {
		var err error
		aTree, err = objectToTree(repo, parent)
		if err != nil {
			return "", err
		}
		defer aTree.Free()
	}

	bTree, err := objectToTree(repo, child)
	if err != nil {
		return "", err
	}
	defer bTree.Free()

	options, err := git.DefaultDiffOptions()
	if err != nil {
		return "", err
	}

	diff, err := repo.DiffTreeToTree(aTree, bTree, &options)
	if err != nil {
		return "", err
	}
	defer diff.Free()

	numDeltas, err := diff.NumDeltas()
	if err != nil {
		return "", err
	}

	var results []string
	for i := 0; i < numDeltas; i++ {
		patch, err := diff.Patch(i)
		if err != nil {
			return "", err
		}
		patchString, err := patch.String()
		if err != nil {
			return "", err
		}
		patch.Free()

		results = append(results, patchString)
	}

	return strings.Join(results, "\n"), nil
}

func (c *client) BranchCredentialCounts(
	logger lager.Logger,
	repositoryPath string,
	sniffer sniff.Sniffer,
	branchType git.BranchType,
) (map[string]uint, error) {
	repo, err := git.OpenRepository(repositoryPath)
	if err != nil {
		return nil, err
	}
	defer repo.Free()

	it, err := repo.NewBranchIterator(branchType)
	if err != nil {
		return nil, err
	}
	defer it.Free()

	var branch *git.Branch
	var target *git.Oid
	var commit *git.Commit
	var tree *git.Tree
	var branchName string
	var blob *git.Blob

	entryCounts := make(map[git.Oid]uint)
	branchCounts := make(map[string]uint)

	for {
		branch, _, err = it.Next()
		if err != nil {
			break
		}

		target = branch.Target()
		if target == nil {
			continue
		}

		commit, err = repo.LookupCommit(target)
		if err != nil {
			return nil, err
		}

		tree, err = commit.Tree()
		if err != nil {
			return nil, err
		}

		branchName, err = branch.Name()
		if err != nil {
			return nil, err
		}

		err = tree.Walk(func(root string, entry *git.TreeEntry) int {
			if entry.Type == git.ObjectBlob {
				if count, ok := entryCounts[*entry.Id]; ok {
					if count > 0 {
						branchCounts[branchName] += count
					}
					return 0
				}

				blob, err = repo.LookupBlob(entry.Id)
				if err != nil {
					return -1
				}

				var count uint
				r := bufio.NewReader(bytes.NewReader(blob.Contents()))
				mime := mimetype.Mimetype(logger, r)
				if mime == "" || strings.HasPrefix(mime, "text") {
					sniffer.Sniff(
						logger,
						filescanner.New(r, entry.Name),
						func(lager.Logger, scanners.Violation) error {
							count++
							return nil
						},
					)
				}

				entryCounts[*entry.Id] = count
				branchCounts[branchName] += count
			}

			return 0
		})

		if err != nil {
			return nil, err
		}
	}

	if blob != nil {
		blob.Free()
	}

	if tree != nil {
		tree.Free()
	}

	if commit != nil {
		commit.Free()
	}

	if branch != nil {
		branch.Free()
	}

	return branchCounts, nil
}

func objectToTree(repo *git.Repository, oid *git.Oid) (*git.Tree, error) {
	object, err := repo.Lookup(oid)
	if err != nil {
		return nil, err
	}
	defer object.Free()

	commit, err := object.AsCommit()
	if err != nil {
		return nil, err
	}
	defer commit.Free()

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func newCredentialsCallback(privateKeyPath, publicKeyPath string) git.CredentialsCallback {
	return func(url string, username string, allowedTypes git.CredType) (git.ErrorCode, *git.Cred) {
		passphrase := ""
		ret, cred := git.NewCredSshKey(username, publicKeyPath, privateKeyPath, passphrase)
		if ret != 0 {
			fmt.Printf("ret: %d\n", ret)
		}
		return git.ErrorCode(ret), &cred
	}
}

func certificateCheckCallback(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
	// should return an error code if the cert isn't valid
	return git.ErrorCode(0)
}

func newFetchOptions(privateKeyPath, publicKeyPath string) *git.FetchOptions {
	credentialsCallback := newCredentialsCallback(privateKeyPath, publicKeyPath)

	return &git.FetchOptions{
		UpdateFetchhead: true,
		RemoteCallbacks: git.RemoteCallbacks{
			CredentialsCallback:      credentialsCallback,
			CertificateCheckCallback: certificateCheckCallback,
		},
	}
}
