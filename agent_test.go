package goblice_test

import (
  "testing"
  "time"
  "log"

  "github.com/xhs/goblice"
)

func TestNewAgent(t *testing.T) {
  agent, err := goblice.NewAgent()
  if err != nil {
    t.Error(err)
  }
  defer agent.Destroy()
}

func TestNewReliableAgent(t *testing.T) {
  agent, err := goblice.NewReliableAgent()
  if err != nil {
    t.Error(err)
  }
  defer agent.Destroy()
}

func TestGenerateCandidates(t *testing.T) {
  agent, _ := goblice.NewAgent()
  defer agent.Destroy()

  if err := agent.GatherCandidates(); err != nil {
    t.Error(err)
  }

  delay := time.After(2 * time.Second)

  for {
    select {
    case <-delay:
      log.Print("timeout")
      break
    case cand := <-agent.Candidates:
      log.Print(cand)
      continue
    case e := <-agent.Events:
      log.Print(e)
      continue
    }

    break
  }
}
