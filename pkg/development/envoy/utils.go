package envoy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	legacyproto "github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

// ToJSONMap converts a proto message to a generic map using canonical JSON encoding
// JSON encoding is specified here: https://developers.google.com/protocol-buffers/docs/proto3#json
func ToJSONMap(msg proto.Message) (map[string]any, error) {
	js, err := ToJSON(msg)
	if err != nil {
		return nil, err
	}

	// Unmarshal from json bytes to go map
	var data map[string]any
	err = json.Unmarshal([]byte(js), &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ToJSON marshals a proto to canonical JSON
func ToJSON(msg proto.Message) (string, error) {
	return ToJSONWithIndent(msg, "")
}

// ToJSONWithIndent marshals a proto to canonical JSON with pretty printed string
func ToJSONWithIndent(msg proto.Message, indent string) (string, error) {
	return ToJSONWithOptions(msg, indent, false)
}

// ToJSONWithOptions marshals a proto to canonical JSON with options to indent and
// print enums' int values
func ToJSONWithOptions(msg proto.Message, indent string, enumsAsInts bool) (string, error) {
	if msg == nil {
		return "", errors.New("unexpected nil message")
	}

	// Marshal from proto to json bytes
	m := jsonpb.Marshaler{Indent: indent, EnumsAsInts: enumsAsInts}
	return m.MarshalToString(legacyproto.MessageV1(msg))
}

// ToYAML marshals a proto to canonical YAML
func ToYAML(msg proto.Message) (string, error) {
	js, err := ToJSON(msg)
	if err != nil {
		return "", err
	}
	yml, err := yaml.JSONToYAML([]byte(js))
	return string(yml), err
}

var keyMatch = regexp.MustCompile(`\"(\w+)\":`)
var wordSplit = regexp.MustCompile(`(\w)([A-Z0-9])`)

type SnakeCaseMarshaller struct {
	Value interface{}
}

func (sc SnakeCaseMarshaller) MarshalJSON() ([]byte, error) {
	s, err := json.Marshal(sc.Value)
	converted := keyMatch.ReplaceAllFunc(
		s,
		func(match []byte) []byte {
			return bytes.ToLower(wordSplit.ReplaceAll(
				match,
				[]byte(`${1}_${2}`),
			))
		},
	)
	return converted, err
}

// MessageToAnyWithError converts from proto message to proto Any
func MessageToAnyWithError(msg proto.Message) (*anypb.Any, error) {
	b, err := marshal(msg)
	if err != nil {
		return nil, err
	}
	return &anypb.Any{
		// nolint: staticcheck
		TypeUrl: "type.googleapis.com/" + string(msg.ProtoReflect().Descriptor().FullName()),
		Value:   b,
	}, nil
}

// MessageToAny converts from proto message to proto Any
func MessageToAny(msg proto.Message) *anypb.Any {
	out, err := MessageToAnyWithError(msg)
	if err != nil {
		klog.Error(fmt.Sprintf("error marshaling Any %s: %v", prototext.Format(msg), err))
		return nil
	}
	return out
}

func marshal(msg proto.Message) ([]byte, error) {
	if vt, ok := msg.(vtStrictMarshal); ok {
		// Attempt to use more efficient implementation
		// "Strict" is the equivalent to Deterministic=true below
		return vt.MarshalVTStrict()
	}
	// If not available, fallback to normal implementation
	return proto.MarshalOptions{Deterministic: true}.Marshal(msg)
}

type vtStrictMarshal interface {
	MarshalVTStrict() ([]byte, error)
}

func IsInjectedPod(pod *corev1.Pod) (bool, *corev1.Container) {
	if pod.Annotations == nil {
		return false, nil
	}

	if _, ok := pod.Annotations[UUIDAnnotation]; ok {
		for i, c := range pod.Spec.Containers {
			if c.Name == EnvoyContainerName {
				return true, &pod.Spec.Containers[i]
			}
		}
	}

	for _, c := range pod.Spec.InitContainers {
		if c.Name == SidecarInitContainerName {
			return true, nil
		}
	}

	return false, nil
}

func IsWebsocketEnabled(pod *corev1.Pod) bool {
	for _, c := range pod.Spec.Containers {
		if strings.HasPrefix(c.Image, "beclab/ws-gateway") {
			return true
		}
	}

	return false
}
