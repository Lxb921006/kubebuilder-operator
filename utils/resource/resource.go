package resource

import (
	"bytes"
	"fmt"
	crdAppV1 "github.com/Lxb921006/kubebuild-go/api/v1"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"text/template"
)

func parseYaml(templateName string, app *crdAppV1.App) ([]byte, error) {
	file, err := os.ReadFile(fmt.Sprintf("internal/controller/template/%s.yaml", templateName))
	if err != nil {
		return []byte{}, err
	}

	tmpl, err := template.New(templateName).Parse(string(file))
	if err != nil {
		return []byte{}, err
	}

	var buf bytes.Buffer

	if err = tmpl.Execute(&buf, app); err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

func NewDeployment(app *crdAppV1.App) (*appsV1.Deployment, error) {
	var dep = new(appsV1.Deployment)
	b, err := parseYaml("deployment", app)
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(b, dep); err != nil {
		return nil, err
	}

	return dep, nil
}

func NewService(app *crdAppV1.App) (*coreV1.Service, error) {
	var svc = new(coreV1.Service)

	b, err := parseYaml("service", app)
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(b, svc); err != nil {
		return nil, err
	}

	return svc, nil

}
