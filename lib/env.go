package lib

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/go-git/go-git/v5"
)

type CommandEnv struct {
	Repo      string
	Branch    string
	User      string
	OS        string
	Container bool
}

func (ce CommandEnv) String() {
	fmt.Printf("CommandEnv:%+v\n", ce)
}

func NewCommandEnv() (*CommandEnv, error) {
	env := &CommandEnv{}
	
	err := FindRepo(env)
	if err != nil {
		return env, err
	}

	container, err := InContainer()
	if err != nil {
		return env, err
	}

	env.Container = container

	env.OS = runtime.GOOS

	if env.User == "" {
		cu, err := user.Current()
		if err == nil {
			env.User = cu.Username
		}
	}

	return env, nil
}

func FindRepo(env *CommandEnv) error {
	path, _ := filepath.Abs(".")
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return err
	}

	config, err := repo.Config()
	if err != nil {
		return err
	}

	origin, ok := config.Remotes["origin"]
	if ok {
		if len(origin.URLs) > 0 {
			env.Repo = origin.URLs[0]
		}
	}

	head, err := repo.Head()
	if err == nil {
		env.Branch = head.Name().String()
	}

	env.User = config.User.Email

	return nil
}

// InContainer looks in the cgroup and mountinfo to see if there are
// signs this is running in a container.
func InContainer() (bool, error) {
	v1Path := "/proc/1/cgroup"
	v1, err := os.Open(v1Path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer v1.Close()

	// Check if the init process (pid 1) has docker in the control group
	scanner := bufio.NewScanner(v1)
	for scanner.Scan() {
		line := scanner.Bytes()
		if bytes.HasPrefix(line, []byte("1:")) && (bytes.Contains(line, []byte("docker")) || bytes.Contains(line, []byte("lxc"))) {
			return true, nil
		}
	}

	return false, nil
}
