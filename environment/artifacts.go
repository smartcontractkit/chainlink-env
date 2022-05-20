package environment

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	clientV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/remotecommand"
)

// Artifacts is an artifacts dumping structure that copies logs and database dumps for all deployed pods
type Artifacts struct {
	env        *Environment
	DBName     string
	podsClient clientV1.PodInterface
}

// NewArtifacts create new artifacts instance for provided environment
func NewArtifacts(env *Environment) (*Artifacts, error) {
	podsClient := env.Client.ClientSet.CoreV1().Pods(env.Out.GetNamespace())
	return &Artifacts{
		env:        env,
		podsClient: podsClient,
	}, nil
}

// DumpTestResult dumps all pods logs and db dump in a separate test dir
func (a *Artifacts) DumpTestResult(testDir string, dbName string) error {
	a.DBName = dbName
	if err := mkdirIfNotExists(testDir); err != nil {
		return err
	}
	if err := a.writePodArtifacts(testDir); err != nil {
		return err
	}
	return nil
}

func (a *Artifacts) writePodArtifacts(testDir string) error {
	log.Info().
		Str("Test", testDir).
		Msg("Writing test artifacts")
	podsList, err := a.podsClient.List(context.Background(), metaV1.ListOptions{})
	if err != nil {
		log.Err(err).
			Str("Namespace", a.env.Out.GetNamespace()).
			Msg("Error retrieving pod list from K8s environment")
		return err
	}
	for _, pod := range podsList.Items {
		log.Info().
			Str("Pod", pod.Name).
			Msg("Writing pod artifacts")
		appName := pod.Labels["app"]
		instance := pod.Labels["instance"]
		appDir := filepath.Join(testDir, fmt.Sprintf("%s_%s", appName, instance))
		if err := mkdirIfNotExists(appDir); err != nil {
			return err
		}
		err = a.writePodLogs(pod, appDir)
		if err != nil {
			log.Err(err).
				Str("Namespace", a.env.Out.GetNamespace()).
				Str("Pod", pod.Name).
				Msg("Error writing logs for pod")
		}
	}
	return nil
}

func (a *Artifacts) dumpDB(pod coreV1.Pod, container coreV1.Container) (string, error) {
	postRequestBase := a.env.Client.ClientSet.CoreV1().RESTClient().Post().
		Namespace(pod.Namespace).Resource("pods").Name(pod.Name).SubResource("exec")
	exportDBRequest := postRequestBase.VersionedParams(
		&coreV1.PodExecOptions{
			Container: container.Name,
			Command:   []string{"/bin/sh", "-c", "pg_dump", a.DBName},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(a.env.Client.RESTConfig, "POST", exportDBRequest.URL())
	if err != nil {
		return "", err
	}
	outBuff, errBuff := &bytes.Buffer{}, &bytes.Buffer{}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  &bytes.Reader{},
		Stdout: outBuff,
		Stderr: errBuff,
		Tty:    false,
	})
	if err != nil || errBuff.Len() > 0 {
		return "", fmt.Errorf("error in dumping DB contents | STDOUT: %v | STDERR: %v", outBuff.String(),
			errBuff.String())
	}
	return outBuff.String(), err
}

func (a *Artifacts) writePostgresDump(podDir string, pod coreV1.Pod, cont coreV1.Container) error {
	dumpContents, err := a.dumpDB(pod, cont)
	if err != nil {
		return err
	}
	logFile, err := os.Create(filepath.Join(podDir, fmt.Sprintf("%s_dump.sql", cont.Name)))
	if err != nil {
		return err
	}
	_, err = logFile.WriteString(dumpContents)
	if err != nil {
		return err
	}
	if err = logFile.Close(); err != nil {
		return err
	}
	return nil
}

func (a *Artifacts) writeContainerLogs(podDir string, pod coreV1.Pod, cont coreV1.Container) error {
	logFile, err := os.Create(filepath.Join(podDir, cont.Name) + ".log")
	if err != nil {
		return err
	}
	podLogRequest := a.podsClient.GetLogs(pod.Name, &coreV1.PodLogOptions{Container: cont.Name})
	podLogs, err := podLogRequest.Stream(context.Background())
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return err
	}
	_, err = logFile.Write(buf.Bytes())
	if err != nil {
		return err
	}

	if err = logFile.Close(); err != nil {
		return err
	}
	if err = podLogs.Close(); err != nil {
		return err
	}
	return nil
}

// Writes logs for each container in a pod
func (a *Artifacts) writePodLogs(pod coreV1.Pod, appDir string) error {
	for _, c := range pod.Spec.Containers {
		log.Info().
			Str("Container", c.Name).
			Msg("Writing container artifacts")
		if err := a.writeContainerLogs(appDir, pod, c); err != nil {
			return err
		}
		if strings.Contains(c.Image, "postgres") {
			if err := a.writePostgresDump(appDir, pod, c); err != nil {
				return err
			}
		}
	}
	return nil
}

func mkdirIfNotExists(dirName string) error {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		if err = os.MkdirAll(dirName, os.ModePerm); err != nil {
			return errors.Wrapf(err, "failed to create directory: %s", dirName)
		}
	}
	return nil
}
