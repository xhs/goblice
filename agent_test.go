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

func TestGenerateOffer(t *testing.T) {
  agent, _ := goblice.NewAgent()
  defer agent.Destroy()

  log.Print(agent.GenerateSdp())
}

func TestIceNegotiation(t *testing.T) {
  client, _ := goblice.NewAgent()
  if err := client.GatherCandidates(); err != nil {
    t.Error(err)
  }

  server, _ := goblice.NewAgent()
  if err := server.GatherCandidates(); err != nil {
    t.Error(err)
  }

  server.ParseSdp(client.GenerateSdp())
  client.ParseSdp(server.GenerateSdp())

  go client.Run()
  go server.Run()

  clientTimeout := time.After(2 * time.Second)

  go func () {
    for {
      select {
      case cand := <-client.Candidates:
        log.Print("client candidate:", cand)
        server.ParseCandidateSdp(cand)
        continue
      case e := <-client.Events:
        if e == goblice.EventNegotiationDone {
          log.Print("client negotiation done")
          client.Send([]byte("hello"))
        }
        continue
      case <-clientTimeout:
        log.Print("client timeout")
        client.Destroy()
        break
      }

      break
    }
  }()

  serverTimeout := time.After(2 * time.Second)

  for {
    select {
    case cand := <-server.Candidates:
      log.Print("server candidate:", cand)
      server.ParseCandidateSdp(cand)
      continue
    case e := <-server.Events:
      if e == goblice.EventNegotiationDone {
        log.Print("server negotiation done")
      }
      continue
    case d := <-server.DataToRead:
      log.Print("server received:", d)
      continue
    case <-serverTimeout:
      log.Print("server timeout")
      server.Destroy()
      break
    }

    break
  }
}
