package model

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	hexPattern = regexp.MustCompile(`[\da-f]+`)
)

type ApplicationId struct {
	Namespace string
	Kind      ApplicationKind
	Name      string
}

func NewApplicationId(ns string, kind ApplicationKind, name string) ApplicationId {
	switch kind {
	case ApplicationKindReplicaSet:
		parts := strings.Split(name, "-")
		if hexPattern.MatchString(parts[len(parts)-1]) {
			kind = ApplicationKindDeployment
			name = strings.Join(parts[:len(parts)-1], "-")
		}
	case ApplicationKindJob:
		parts := strings.Split(name, "-")
		if _, err := strconv.ParseUint(parts[len(parts)-1], 10, 64); err == nil {
			kind = ApplicationKindCronJob
			name = strings.Join(parts[:len(parts)-1], "-")
		}
	case "", "<none>":
		kind = ApplicationKindPod
	}
	if ns == "" {
		ns = "_"
	}
	return ApplicationId{Namespace: ns, Kind: kind, Name: name}
}

func NewApplicationIdFromString(src string) (ApplicationId, error) {
	parts := strings.SplitN(src, ":", 3)
	if len(parts) < 3 {
		return ApplicationId{}, fmt.Errorf("should be ns:kind:name")
	}
	return NewApplicationId(parts[0], ApplicationKind(parts[1]), parts[2]), nil
}

func (a ApplicationId) String() string {
	return fmt.Sprintf("%s:%s:%s", a.Namespace, a.Kind, a.Name)
}

func (a ApplicationId) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}
