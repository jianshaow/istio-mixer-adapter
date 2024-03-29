// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// THIS FILE IS AUTOMATICALLY GENERATED.

package enhencedauthz

import (
	"context"

	"istio.io/istio/mixer/pkg/adapter"
)

// The `enhencedauthz` template defines parameters for performing policy
// enforcement within Istio. It is primarily concerned with enabling Mixer

// Fully qualified name of the template
const TemplateName = "enhencedauthz"

// Instance is constructed by Mixer for the 'enhencedauthz' template.
//
// The `enhencedauthz` template defines parameters for performing policy
// enforcement within Istio. It is primarily concerned with enabling Mixer
// adapters to make decisions about who is allowed to do what.
// In this template, the "who" is defined in a Subject message. The "what" is
// defined in an Action message. During a Mixer Check call, these values
// will be populated based on configuration from request attributes and
// passed to individual authorization adapters to adjudicate.
type Instance struct {
	// Name of the instance as specified in configuration.
	Name string

	// A subject contains a list of attributes that identify
	// the caller identity.
	Subject *Subject

	// An action defines "how a resource is accessed".
	Action *Action
}

// Output struct is returned by the attribute producing adapters that handle this template.
//
// The `enhencedauthz` output template defines authoriztion context information will be
// returned to mixer, those can be used in rule.
type Output struct {
	fieldsSet map[string]bool

	// The clientID the ID of the client call the service.
	ClientID string

	// The AuthzType the type of authorization in header.
	AuthzType string

	// A authzContext contains authorization related informations.
	Properties map[string]string
}

func NewOutput() *Output {
	return &Output{fieldsSet: make(map[string]bool)}
}

func (o *Output) SetClientID(val string) {
	o.fieldsSet["clientID"] = true
	o.ClientID = val
}

func (o *Output) SetAuthzType(val string) {
	o.fieldsSet["authzType"] = true
	o.AuthzType = val
}

func (o *Output) SetProperties(val map[string]string) {
	o.fieldsSet["properties"] = true
	o.Properties = val
}

func (o *Output) WasSet(field string) bool {
	_, found := o.fieldsSet[field]
	return found
}

// A subject contains a list of attributes that identify
// the caller identity.
type Subject struct {

	// The user name/ID that the subject represents.
	User string

	// Groups the subject belongs to depending on the authentication mechanism,
	// "groups" are normally populated from JWT claim or client certificate.
	// The operator can define how it is populated when creating an instance of
	// the template.
	Groups string

	// Additional attributes about the subject.
	Properties map[string]interface{}
}

// An action defines "how a resource is accessed".
type Action struct {

	// Namespace the target action is taking place in.
	Namespace string

	// The Service the action is being taken on.
	Service string

	// What action is being taken.
	Method string

	// HTTP REST path within the service
	Path string

	// Additional data about the action for use in policy.
	Properties map[string]interface{}
}

// HandlerBuilder must be implemented by adapters if they want to
// process data associated with the 'enhencedauthz' template.
//
// Mixer uses this interface to call into the adapter at configuration time to configure
// it with adapter-specific configuration as well as all template-specific type information.
type HandlerBuilder interface {
	adapter.HandlerBuilder

	// SetEnhencedauthzTypes is invoked by Mixer to pass the template-specific Type information for instances that an adapter
	// may receive at runtime. The type information describes the shape of the instance.
	SetEnhencedauthzTypes(map[string]*Type /*Instance name -> Type*/)
}

// Handler must be implemented by adapter code if it wants to
// process data associated with the 'enhencedauthz' template.
//
// Mixer uses this interface to call into the adapter at request time in order to dispatch
// created instances to the adapter. Adapters take the incoming instances and do what they
// need to achieve their primary function.
//
// The name of each instance can be used as a key into the Type map supplied to the adapter
// at configuration time via the method 'SetEnhencedauthzTypes'.
// These Type associated with an instance describes the shape of the instance
type Handler interface {
	adapter.Handler

	// HandleEnhencedauthz is called by Mixer at request time to deliver instances to
	// to an adapter.
	HandleEnhencedauthz(context.Context, *Instance) (adapter.CheckResult, *Output, error)
}
