package goblice

/*
#include <stdlib.h>
#include <nice/agent.h>

extern void go_candidate_gathering_done_cb(NiceAgent *agent, guint stream, gpointer udata);
static void set_candidate_gathering_done_cb(NiceAgent *agent, gpointer udata) {
    g_signal_connect(G_OBJECT(agent), "candidate-gathering-done",
                     G_CALLBACK(go_candidate_gathering_done_cb), udata);
}

extern void go_component_state_changed_cb(NiceAgent *agent, guint stream, 
                                          guint component, guint state, gpointer udata);
static void set_component_state_changed_cb(NiceAgent *agent, gpointer udata) {
    g_signal_connect(G_OBJECT(agent), "component-state-changed",
                     G_CALLBACK(go_component_state_changed_cb), udata);
}

extern void go_new_candidate_cb(NiceAgent *agent, NiceCandidate *candidate, gpointer udata);
static void set_new_candidate_cb(NiceAgent *agent, gpointer udata) {
  g_signal_connect(G_OBJECT(agent), "new-candidate-full",
                   G_CALLBACK(go_new_candidate_cb), udata);
}

extern void go_new_selected_pair_cb(NiceAgent *agent, guint stream, guint component,
                                    NiceCandidate *lcand, NiceCandidate *rcand, gpointer udata);
static void set_new_selected_pair_cb(NiceAgent *agent, gpointer udata) {
  g_signal_connect(G_OBJECT(agent), "new-selected-pair-full",
                   G_CALLBACK(go_new_selected_pair_cb), udata);
}
*/
import "C"

import (
  "unsafe"
)

type CallbackData struct {
  cb interface{}
  data interface{}
}

type CandidateGatheringDoneCB func(a *Agent, stream int, udata interface{})

//export go_candidate_gathering_done_cb
func go_candidate_gathering_done_cb(agent *C.NiceAgent, stream C.guint, udata unsafe.Pointer) {
  d := (*CallbackData)(udata)
  a := &Agent{agent: agent}
  d.cb.(CandidateGatheringDoneCB)(a, int(stream), d.data)
}

func (a *Agent) SetCandidateGatheringDoneCB(cb CandidateGatheringDoneCB, udata interface{}) error {
  d := unsafe.Pointer(&CallbackData{cb, udata})
  C.set_candidate_gathering_done_cb(a.agent, C.gpointer(d))
  return nil
}

type ComponentStateChangedCB func(a *Agent, stream, component, state int, udata interface{})

//export go_component_state_changed_cb
func go_component_state_changed_cb(agent *C.NiceAgent, stream, component, state C.guint, udata unsafe.Pointer) {
  d := (*CallbackData)(udata)
  a := &Agent{agent: agent}
  d.cb.(ComponentStateChangedCB)(a, int(stream), int(component), int(state), d.data)
}

func (a *Agent) SetComponentStateChangedCB(cb ComponentStateChangedCB, udata interface{}) error {
  d := unsafe.Pointer(&CallbackData{cb, udata})
  C.set_component_state_changed_cb(a.agent, C.gpointer(d))
  return nil
}

type NewCandidateCB func(a *Agent, c *Candidate, udata interface{})

//export go_new_candidate_cb
func go_new_candidate_cb(agent *C.NiceAgent, candidate *C.NiceCandidate, udata unsafe.Pointer) {
  d := (*CallbackData)(udata)
  a := &Agent{agent: agent}
  c := &Candidate{candidate}
  d.cb.(NewCandidateCB)(a, c, d.data)
}

func (a *Agent) SetNewCandidateCB(cb NewCandidateCB, udata interface{}) error {
  d := unsafe.Pointer(&CallbackData{cb, udata})
  C.set_new_candidate_cb(a.agent, C.gpointer(d))
  return nil
}

type NewSelectedPairCB func(a *Agent, stream, component int, l, r *Candidate, udata interface{})

//export go_new_selected_pair_cb
func go_new_selected_pair_cb(agent *C.NiceAgent, stream, component C.guint,
                             lcand, rcand *C.NiceCandidate, udata unsafe.Pointer) {
  d := (*CallbackData)(udata)
  a := &Agent{agent: agent}
  l := &Candidate{lcand}
  r := &Candidate{rcand}
  d.cb.(NewSelectedPairCB)(a, int(stream), int(component), l, r, d.data)
}

func (a *Agent) SetNewSelectedPairCB(cb NewSelectedPairCB, udata interface{}) error {
  d := unsafe.Pointer(&CallbackData{cb, udata})
  C.set_new_selected_pair_cb(a.agent, C.gpointer(d))
  return nil
}
