package podman

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	corev1 "k8s.io/api/core/v1"
	jsonserializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/klog"
	"k8s.io/kubectl/pkg/scheme"
)

type PodmanCli struct{}

func NewPodmanCli() *PodmanCli {
	return &PodmanCli{}
}

func (o *PodmanCli) PlayKube(pod *corev1.Pod) error {
	serializer := jsonserializer.NewSerializerWithOptions(
		jsonserializer.SimpleMetaFactory{},
		scheme.Scheme,
		scheme.Scheme,
		jsonserializer.SerializerOptions{
			Yaml: true,
		},
	)

	cmd := exec.Command("podman", "play", "kube", "-")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout

	if err = cmd.Start(); err != nil {
		return err
	}

	err = serializer.Encode(pod, stdin)
	if err != nil {
		return err
	}
	stdin.Close()

	go func() {
		for {
			tmp := make([]byte, 1024)
			_, err = stdout.Read(tmp)
			klog.V(4).Info(string(tmp))
			if err != nil {
				break
			}
		}
	}()
	if err = cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return err
	}

	return nil
}

func (o *PodmanCli) PodStop(podname string) error {
	out, err := exec.Command("podman", "pod", "stop", podname).Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return err
	}
	klog.V(4).Infof("Stopped pod %s", string(out))
	return nil
}

func (o *PodmanCli) PodRm(podname string) error {
	out, err := exec.Command("podman", "pod", "rm", podname).Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return err
	}
	klog.V(4).Infof("Deleted pod %s", string(out))
	return nil
}

func (o *PodmanCli) VolumeRm(volumeName string) error {
	out, err := exec.Command("podman", "volume", "rm", volumeName).Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return err
	}
	klog.V(4).Infof("Deleted volume %s", string(out))
	return nil
}

func (o *PodmanCli) VolumeLs() (map[string]bool, error) {
	out, err := exec.Command("podman", "volume", "ls", "--format", "{{.Name}}", "--noheading").Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return nil, err
	}
	return SplitLinesAsSet(string(out)), nil
}

func SplitLinesAsSet(s string) map[string]bool {
	lines := map[string]bool{}
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		lines[sc.Text()] = true
	}
	return lines
}
