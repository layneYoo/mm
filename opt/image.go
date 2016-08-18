package opt

import (
	"fmt"
	"github.com/layneYoo/mCtl/check"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
)

type DeployItem struct {
	Version     string
	Appname     string
	Constraints string
}

type ImageBuild struct {
	//gitDstPath string
}

func (m ImageBuild) Apply(args []string) {
	check.Check(len(args) == 5, "four arguments needed")
	if args[0] == "" || args[1] == "" || args[2] == "" {
		log.Fatal("argument null")
		return
	}
	buildPath := args[0]
	registryPath := args[1]
	gitUrl := args[2]
	deployTpl := args[3]
	TlpnamePre := strings.Split(path.Base(deployTpl), ".")[0]
	TlpPath := path.Dir(args[3])
	dockerConstraints := args[4]
	testCts := strings.Split(strings.Split(dockerConstraints, ";")[0], ",")
	proCts := strings.Split(strings.Split(dockerConstraints, ";")[1], ",")
	// testing the path
	_, err := os.Stat(buildPath)
	if err != nil {
		existOr := os.IsExist(err)
		//Check(existOr, "error : ["+buildPath+"] No such directory")
		// not exist, git clone
		if existOr == false {
			out, err := exec.Command("bash", "-c", "/usr/local/bin/git clone "+gitUrl+" "+buildPath).Output()
			check.Check(err == nil, "git clone error")
			fmt.Println("\n git clone " + string(out))
		}
	} else {
		// exist, git pull
		out, err := exec.Command("bash", "-c", "cd "+buildPath+" && /usr/local/bin/git pull origin master").Output()
		check.Check(err == nil, "git pull error")
		fmt.Println("\n " + string(out))
	}
	// testing registryPath
	// testing the json
	buildCmdHead := "cd "
	buildCmdGitV := ` && /usr/local/bin/git log -1 | head -1 | awk -F" " '{print $2}'`

	// get the commit version
	out, err := exec.Command("bash", "-c", buildCmdHead+buildPath+buildCmdGitV).Output()
	check.Check(err == nil, "get git version error")
	//gitVersion := string(out[0 : len(out)-2])
	gitVersion := string(out[0:9])

	// build docker image
	buildCmdBuild := ` && docker build -t `
	out, err = exec.Command("bash", "-c", buildCmdHead+buildPath+buildCmdBuild+registryPath+":"+gitVersion+" .").Output()
	check.Check(err == nil, "build command error:")
	fmt.Println("\n" + string(out))

	// create the marathon's json for deploying
	tlp, err := template.ParseFiles(deployTpl)
	check.Check(err == nil, "template parsefile error")
	deployTestCts := "\""
	for i := 0; i < len(testCts); i++ {
		if i < len(testCts)-1 {
			deployTestCts += testCts[i] + "\", \""
		} else {
			deployTestCts += testCts[i]
		}
	}
	deployTestCts += "\""
	deployOnlineCts := "\""
	for i := 0; i < len(proCts); i++ {
		if i < len(proCts)-1 {
			deployOnlineCts += proCts[i] + "\", \""
		} else {
			deployOnlineCts += proCts[i]
		}
	}
	deployOnlineCts += "\""
	deployTest := DeployItem{Version: gitVersion, Appname: TlpnamePre + "-test", Constraints: deployTestCts}
	deployOnline := DeployItem{Version: gitVersion, Appname: TlpnamePre + "-Online", Constraints: deployOnlineCts}
	deployNameTest := TlpPath + "/" + TlpnamePre + "_test.json"
	deployNameOnline := TlpPath + "/" + TlpnamePre + "_pro.json"
	ofpTest, err := os.OpenFile(deployNameTest, os.O_WRONLY|os.O_CREATE, 0666)
	ofpOnline, err := os.OpenFile(deployNameOnline, os.O_WRONLY|os.O_CREATE, 0666)
	check.Check(err == nil, "create file error")
	defer ofpTest.Close()
	defer ofpOnline.Close()
	err = tlp.Execute(ofpOnline, deployOnline)
	check.Check(err == nil, "template pro execute error")
	err = tlp.Execute(ofpTest, deployTest)
	check.Check(err == nil, "template test execute error")
}

type ImageUpload struct {
}

func (m ImageUpload) Apply(args []string) {
	check.Check(len(args) == 2, "two arguments needed")
	if args[0] == "" || args[1] == "" {
		log.Fatal("argument null")
		return
	}
	buildPath := args[0]
	registryPath := args[1]
	buildCmdHead := "cd "
	buildCmdGitV := ` && /usr/local/bin/git log -1 | head -1 | awk -F" " '{print $2}'`

	// get the commit version
	out, err := exec.Command("bash", "-c", buildCmdHead+buildPath+buildCmdGitV).Output()
	check.Check(err == nil, "get git version error")
	gitVersion := string(out[0:9])

	// push the image[ registryPath:gitVersion ]
	out, err = exec.Command("bash", "-c", "docker push "+registryPath+":"+gitVersion).Output()
	check.Check(err == nil, "get git version error")
	fmt.Println(string(out))
}
