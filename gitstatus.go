package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	// change those symbols to whatever you prefer
	symbols := map[string]string{"ahead of": "↑", "behind": "↓", "prehash": ":"}

	gitsym, err := exec.Command("git", "symbolic-ref", "HEAD").CombinedOutput()
	if strings.Contains(string(gitsym), "fatal: Not a git repository") {
		os.Exit(0)
	}
	var branch string
	if strings.Contains(string(gitsym), "fatal: ref HEAD is not a symbolic ref") {
		hash, _ := exec.Command(
			"git",
			"rev-parse",
			"--short",
			"HEAD").Output()
		branch = symbols["prehash"] + string(hash[:len(hash)-1])
	} else {
		branch = strings.TrimSpace(string(gitsym))[11:]
	}

	res, err := exec.Command("git", "diff", "--name-status").Output()
	if err != nil {
		os.Exit(0)
	}
	var changedFiles []string
	for _, namestat := range strings.Split(string(res), "\n") {
		if len(namestat) > 0 {
			changedFiles = append(changedFiles, string(namestat[0]))
		}
	}

	res, _ = exec.Command("git", "diff", "--staged", "--name-status").Output()
	var stagedFiles []string
	for _, namestat := range strings.Split(string(res), "\n") {
		if len(namestat) > 0 {
			stagedFiles = append(stagedFiles, string(namestat[0]))
		}
	}

	nbChanged := len(changedFiles) -
		strings.Count(strings.Join(changedFiles, ""), "U")
	nbU := strings.Count(strings.Join(stagedFiles, ""), "U")
	nbStaged := len(stagedFiles) - nbU
	staged := strconv.Itoa(nbStaged)
	conflicts := strconv.Itoa(nbU)
	changed := strconv.Itoa(nbChanged)

	res, _ = exec.Command(
		"git",
		"ls-files",
		"--others",
		"--exclude-standard").Output()
	var untrackedFiles []string
	for _, namestat := range strings.Split(string(res), "\n") {
		if len(namestat) > 0 {
			untrackedFiles = append(untrackedFiles, string(namestat[0]))
		}
	}

	nbUntracked := len(untrackedFiles)
	untracked := strconv.Itoa(nbUntracked)

	var clean string
	if nbChanged+nbStaged+nbU+nbUntracked == 0 {
		clean = "1"
	} else {
		clean = "0"
	}

	var remote string

	if len(branch) == 0 { // not on any branch
		branch = symbols["prehash"]
	} else {
		remoteParam := fmt.Sprintf("branch.%s.remote", branch)
		remoteBytes, _ := exec.Command("git", "config", remoteParam).Output()
		remoteName := strings.TrimSpace(string(remoteBytes))
		if len(remoteName) > 0 {
			mergeParam := fmt.Sprintf("branch.%s.merge", branch)
			mergeBytes, _ := exec.Command("git", "config", mergeParam).Output()
			mergeName := strings.TrimSpace(string(mergeBytes))
			var remoteRef string
			if remoteName == "." { // local
				remoteRef = mergeName
			} else {
				remoteRef = fmt.Sprintf(
					"refs/remotes/%s/%s",
					remoteName,
					mergeName[11:])
			}
			revParam := fmt.Sprintf("%s...HEAD", remoteRef)
			revlist, err := exec.Command(
				"git",
				"rev-list",
				"--left-right",
				revParam).Output()
			if err != nil { // fallback to local
				revParam = fmt.Sprintf("%s...HEAD", mergeName)
				revlist, _ = exec.Command(
					"git",
					"rev-list",
					"--left-right",
					revParam).Output()
			}
			behead := strings.Split(string(revlist), "\n")
			var ahead int
			for _, x := range behead {
				if len(x) > 0 {
					if string(x[0]) == ">" {
						ahead++
					}
				}
			}
			behind := len(behead) - 1 - ahead
			if behind != 0 {
				remote = fmt.Sprintf("%s%d", symbols["behind"], behind)
			}
			if ahead != 0 {
				remote += fmt.Sprintf("%s%d", symbols["ahead of"], ahead)
			}
		}

	}

	out := strings.Join([]string{
		branch,
		remote,
		staged,
		conflicts,
		changed,
		untracked,
		clean}, "\n")
	fmt.Println(out)
}
