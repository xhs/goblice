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

extern void go_data_received_cb(NiceAgent *agent, guint stream_id, guint component_id,
  						   	    					guint len, gchar *buf, gpointer udata);
static void attach_receive_db(NiceAgent *agent, guint stream_id, GMainLoop *loop, void *udata) {
	nice_agent_attach_recv(agent, stream_id, 1, g_main_loop_get_context(loop), go_data_received_cb, udata);
}

extern void go_candidate_gathering_done_cb(NiceAgent *agent, guint stream, gpointer udata);

extern void go_new_candidate_cb(NiceAgent *agent, NiceCandidate *candidate, gpointer udata);

extern void go_new_selected_pair_cb(NiceAgent *agent, guint stream, guint component,
                                    NiceCandidate *lcand, NiceCandidate *rcand, gpointer udata);

static void set_callbacks(NiceAgent *agent, void *udata) {
	g_signal_connect(G_OBJECT(agent), "candidate-gathering-done",
                   G_CALLBACK(go_candidate_gathering_done_cb), udata);
	g_signal_connect(G_OBJECT(agent), "new-candidate-full",
                   G_CALLBACK(go_new_candidate_cb), udata);
  g_signal_connect(G_OBJECT(agent), "new-selected-pair-full",
                   G_CALLBACK(go_new_selected_pair_cb), udata);
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

const (
	EventGatheringDone = 0
	EventNegotiationDone = 1
)

type Agent struct {
	agent *C.NiceAgent
	loop  *C.GMainLoop
	stream int
	DataToRead chan []byte
	Events chan int
	Candidates chan string
}

type Candidate struct {
	cand *C.NiceCandidate
}

//export go_candidate_gathering_done_cb
func go_candidate_gathering_done_cb(agent *C.NiceAgent, stream C.guint, udata unsafe.Pointer) {
	a := (*Agent)(udata)
	a.Events <- EventGatheringDone
}

//export go_new_candidate_cb
func go_new_candidate_cb(agent *C.NiceAgent, candidate *C.NiceCandidate, udata unsafe.Pointer) {
	s := C.nice_agent_generate_local_candidate_sdp(agent, candidate)
	defer C.free(unsafe.Pointer(s))
	c := C.GoString((*C.char)(s))
	a := (*Agent)(udata)
	a.Candidates <- c
}

//export go_new_selected_pair_cb
func go_new_selected_pair_cb(agent *C.NiceAgent, stream, component C.guint,
	lcand, rcand *C.NiceCandidate, udata unsafe.Pointer) {
	a := (*Agent)(udata)
	a.Events <- EventNegotiationDone
}

//export go_data_received_cb
func go_data_received_cb(agent *C.NiceAgent, stream, component, length C.guint,
	data *C.gchar, udata unsafe.Pointer) {
	a := (*Agent)(udata)
	b := C.GoBytes(unsafe.Pointer(data), C.int(length))
	a.DataToRead <- b
}

func NewAgent() (*Agent, error) {
	return newAgent(false)
}

func NewReliableAgent() (*Agent, error) {
	return newAgent(true)
}

func newAgent(reliable bool) (*Agent, error) {
	var loop *C.GMainLoop
	loop = C.g_main_loop_new(nil, 0)
	if loop == nil {
		return nil, errors.New("failed to create new main loop")
	}

	var agent *C.NiceAgent
	if reliable {
		agent = C.nice_agent_new_reliable(C.g_main_loop_get_context(loop), C.NICE_COMPATIBILITY_RFC5245)
	} else {
		agent = C.nice_agent_new(C.g_main_loop_get_context(loop), C.NICE_COMPATIBILITY_RFC5245)
	}
	if agent == nil {
		C.g_main_loop_unref(loop)
		return nil, errors.New("failed to create new agent")
	}

	cs := C.CString("controlling-mode")
	defer C.free(unsafe.Pointer(cs))
	C.g_object_set_int_wrap(C.gpointer(agent), cs, 1)

	stream := C.nice_agent_add_stream(agent, 1)
	if stream == 0 {
		C.g_main_loop_unref(loop)
		C.g_object_unref(C.gpointer(agent))
		return nil, errors.New("failed to add stream")
	}

	ns := C.CString("application")
	defer C.free(unsafe.Pointer(ns))
	rv := C.nice_agent_set_stream_name(agent, stream, (*C.gchar)(ns))
	if rv == 0 {
		C.g_main_loop_unref(loop)
		C.g_object_unref(C.gpointer(agent))
		return nil, errors.New("failed to set stream name")
	}

	a := &Agent{agent: agent, loop: loop, stream: int(stream)}
	a.DataToRead = make(chan []byte, 16)
	a.Events = make(chan int, 2)
	a.Candidates = make(chan string, 16)
	C.attach_receive_db(agent, stream, loop, unsafe.Pointer(a))
	C.set_callbacks(agent, unsafe.Pointer(a))

	return a, nil
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

func (a *Agent) SetStunServer(ip string) {
	s := C.CString("stun-server")
	defer C.free(unsafe.Pointer(s))
	v := C.CString(ip)
	defer C.free(unsafe.Pointer(v))
	C.g_object_set_string_wrap(C.gpointer(a.agent), s, v)
}

func (a *Agent) SetStunPort(port int) {
	s := C.CString("stun-server-port")
	defer C.free(unsafe.Pointer(s))
	C.g_object_set_int_wrap(C.gpointer(a.agent), s, C.int(port))
}

func (a *Agent) GatherCandidates() error {
	rv := int(C.nice_agent_gather_candidates(a.agent, C.guint(a.stream)))
	if rv == 0 {
		return errors.New("failed to gather candidates")
	}
	return nil
}

func (a *Agent) Send(data []byte) (int, error) {
	tv := C.nice_agent_send(a.agent, C.guint(a.stream), 1,
		C.guint(len(data)), (*C.gchar)(unsafe.Pointer(&data[0])))
	if tv < 0 {
		return 0, errors.New("failed to send data")
	}
	return int(tv), nil
}

func (a *Agent) GenerateSdp() string {
	s := C.nice_agent_generate_local_sdp(a.agent)
	defer C.free(unsafe.Pointer(s))
	return C.GoString((*C.char)(s))
}

func (a *Agent) GenerateCandidateSdp(c *Candidate) string {
	s := C.nice_agent_generate_local_candidate_sdp(a.agent, c.cand)
	defer C.free(unsafe.Pointer(s))
	return C.GoString((*C.char)(s))
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

func (a *Agent) ParseCandidateSdp(sdp string) (*Candidate, error) {
	s := C.CString(sdp)
	defer C.free(unsafe.Pointer(s))
	c := C.nice_agent_parse_remote_candidate_sdp(a.agent, C.guint(a.stream), (*C.gchar)(s))
	if c == nil {
		return nil, errors.New("invalid remote candidate sdp")
	}
	return &Candidate{c}, nil
}

func (a *Agent) AddRemoteCandidate(c *Candidate) (int, error) {
	list := C.g_slist_append(nil, C.gpointer(c.cand))
	defer C.g_slist_free(list)
	rv := C.nice_agent_set_remote_candidates(a.agent, C.guint(a.stream), 1, list)
	if rv < 0 {
		return 0, errors.New("failed to add remote candidate")
	}
	return int(rv), nil
}
