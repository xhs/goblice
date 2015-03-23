package goblice

/*
#cgo pkg-config: nice

#include <stdlib.h>
#include <nice/agent.h>

static void g_object_set_int_wrap(gpointer object, const char *property, int value) {
  g_object_set(object, property, value, NULL);
}

static void g_object_set_string_wrap(gpointer object, const char *property, const char *value) {
  g_object_set(object, property, value, NULL);
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

type Agent struct {
	agent *C.NiceAgent
	loop  *C.GMainLoop
}

type Candidate struct {
	cand *C.NiceCandidate
}

func NewAgent() (*Agent, error) {
	var loop *C.GMainLoop
	loop = C.g_main_loop_new(nil, 0)
	if loop == nil {
		return nil, errors.New("failed to create new main loop")
	}

	var agent *C.NiceAgent
	agent = C.nice_agent_new(C.g_main_loop_get_context(loop), C.NICE_COMPATIBILITY_RFC5245)
	if agent == nil {
		return nil, errors.New("failed to create new agent")
	}

	return &Agent{agent, loop}, nil
}

func NewReliableAgent() (*Agent, error) {
	var loop *C.GMainLoop
	loop = C.g_main_loop_new(nil, 0)
	if loop == nil {
		return nil, errors.New("failed to create new main loop")
	}

	var agent *C.NiceAgent
	agent = C.nice_agent_new_reliable(C.g_main_loop_get_context(loop), C.NICE_COMPATIBILITY_RFC5245)
	if agent == nil {
		return nil, errors.New("failed to create new agent")
	}

	return &Agent{agent, loop}, nil
}

func (a *Agent) Run() error {
	C.g_main_loop_run(a.loop)
	return nil
}

func (a *Agent) Destroy() error {
	C.g_main_loop_quit(a.loop)
	C.g_object_unref(C.gpointer(a.agent))
	C.g_main_loop_unref(a.loop)
	return nil
}

func (a *Agent) SetControllingMode(controlling int) error {
	s := C.CString("controlling-mode")
	defer C.free(unsafe.Pointer(s))
	C.g_object_set_int_wrap(C.gpointer(a.agent), s, C.int(controlling))
	return nil
}

func (a *Agent) SetStunServer(ip string) error {
	s := C.CString("stun-server")
	defer C.free(unsafe.Pointer(s))
	v := C.CString(ip)
	defer C.free(unsafe.Pointer(v))
	C.g_object_set_string_wrap(C.gpointer(a.agent), s, v)
	return nil
}

func (a *Agent) SetStunServerPort(port int) error {
	s := C.CString("stun-server-port")
	defer C.free(unsafe.Pointer(s))
	C.g_object_set_int_wrap(C.gpointer(a.agent), s, C.int(port))
	return nil
}

func (a *Agent) AddStream(components int) (int, error) {
	rv := int(C.nice_agent_add_stream(a.agent, C.guint(components)))
	if rv == 0 {
		return 0, errors.New("failed to add stream")
	}
	return rv, nil
}

func (a *Agent) SetStreamName(stream int, name string) error {
	s := C.CString(name)
	defer C.free(unsafe.Pointer(s))
	rv := int(C.nice_agent_set_stream_name(a.agent, C.guint(stream), (*C.gchar)(s)))
	if rv == 0 {
		return errors.New("failed to set stream name")
	}
	return nil
}

func (a *Agent) GatherCandidates(stream int) error {
	rv := int(C.nice_agent_gather_candidates(a.agent, C.guint(stream)))
	if rv == 0 {
		return errors.New("failed to gather candidates")
	}
	return nil
}

func (a *Agent) Send(stream, component, length int, buf *[]byte) (int, error) {
	tv := C.nice_agent_send(a.agent, C.guint(stream), C.guint(component),
		C.guint(length), (*C.gchar)(unsafe.Pointer(buf)))
	if tv < 0 {
		return 0, errors.New("failed to send data")
	}
	return int(tv), nil
}

func (a *Agent) Receive(stream, component, length int, buf *[]byte) (int, error) {
	rv := C.nice_agent_recv(a.agent, C.guint(stream), C.guint(component),
		(*C.guint8)(unsafe.Pointer(buf)), C.gsize(length), nil, nil)
	if rv < 0 {
		return 0, errors.New("failed to receive data")
	}
	return int(rv), nil
}

func (a *Agent) GenerateSdp() (string, error) {
	s := C.nice_agent_generate_local_sdp(a.agent)
	defer C.free(unsafe.Pointer(s))
	return C.GoString((*C.char)(s)), nil
}

func (a *Agent) ParseSdp(sdp string) (int, error) {
	s := C.CString(sdp)
	defer C.free(unsafe.Pointer(s))
	rv := C.nice_agent_parse_remote_sdp(a.agent, (*C.gchar)(s))
	if rv < 0 {
		return 0, errors.New("invalid remote sdp")
	}
	return int(rv), nil
}

func (a *Agent) ParseCandidateSdp(stream int, sdp string) (*Candidate, error) {
	s := C.CString(sdp)
	defer C.free(unsafe.Pointer(s))
	c := C.nice_agent_parse_remote_candidate_sdp(a.agent, C.guint(stream), (*C.gchar)(s))
	if c == nil {
		return nil, errors.New("invalid remote candidate sdp")
	}
	return &Candidate{c}, nil
}

func (a *Agent) AddRemoteCandidate(stream int, c *Candidate) (int, error) {
	list := C.g_slist_append(nil, C.gpointer(c.cand))
	defer C.g_slist_free(list)
	rv := C.nice_agent_set_remote_candidates(a.agent, C.guint(stream), 1, list)
	if rv < 0 {
		return 0, errors.New("failed to add remote candidate")
	}
	return int(rv), nil
}
