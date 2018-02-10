package treesql

// maybe this should be in a different package idk
// this should pretty much be the same API as TreeSQLClient.js

import (
	"errors"

	"log"

	"github.com/gorilla/websocket"
)

type ClientConn struct {
	WebSocketConn    *websocket.Conn
	NextStatementID  int
	StatementsToSend chan *StatementRequest
	IncomingMessages chan *ChannelMessage
	Channels         map[int]*ClientChannel
}

type StatementRequest struct {
	Statement  string
	ResultChan chan *ClientChannel
}

func NewClientConn(url string) (*ClientConn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	clientConn := &ClientConn{
		NextStatementID:  0,
		WebSocketConn:    conn,
		StatementsToSend: make(chan *StatementRequest),
		IncomingMessages: make(chan *ChannelMessage),
		Channels:         map[int]*ClientChannel{},
	}
	go clientConn.handleStatements()
	go clientConn.handleIncoming()
	return clientConn, nil
}

func (conn *ClientConn) Close() error {
	return conn.WebSocketConn.Close()
	// idk if it should also do something to the channels
}

func (conn *ClientConn) handleStatements() {
	for {
		select {
		case request := <-conn.StatementsToSend:
			channel := &ClientChannel{
				Conn:        conn,
				StatementID: conn.NextStatementID,
				Statement:   request.Statement,
				Updates:     make(chan *MessageToClient),
			}
			conn.NextStatementID++
			conn.Channels[channel.StatementID] = channel
			request.ResultChan <- channel
			conn.WebSocketConn.WriteMessage(websocket.TextMessage, []byte(request.Statement))

		case incomingMsg := <-conn.IncomingMessages:
			channel := conn.Channels[incomingMsg.StatementID]
			channel.Updates <- incomingMsg.Message
		}
	}
}

func (conn *ClientConn) handleIncoming() {
	defer conn.WebSocketConn.Close()
	for {
		parsedMessage := &ChannelMessage{}
		err := conn.WebSocketConn.ReadJSON(&parsedMessage)
		if err != nil {
			log.Println("error in handleIncoming:", err)
			// uh... should probably recover gracefully from this, but
			// idk how to return an error from a goroutine. how would its
			// supervisor (???) handle it? I want erlang lol
		}
		conn.IncomingMessages <- parsedMessage
	}
}

type ClientChannel struct {
	Conn        *ClientConn
	StatementID int
	Statement   string
	Updates     chan *MessageToClient
}

func (conn *ClientConn) sendStatement(statement string) *ClientChannel {
	resultChan := make(chan *ClientChannel)
	conn.StatementsToSend <- &StatementRequest{
		ResultChan: resultChan,
		Statement:  statement,
	}
	return <-resultChan
}

func (conn *ClientConn) LiveQuery(query string) *ClientChannel {
	return conn.sendStatement(query)
}

func (conn *ClientConn) Query(query string) (*InitialResult, error) {
	resultChan := conn.sendStatement(query)
	update := <-resultChan.Updates
	if update.ErrorMessage != nil {
		return nil, errors.New(*update.ErrorMessage)
	} else if update.InitialResultMessage != nil {
		return update.InitialResultMessage, nil
	}
	panic("Query result neither error or initial result")
}

func (conn *ClientConn) Exec(statement string) (string, error) {
	resultChan := conn.sendStatement(statement)
	update := <-resultChan.Updates
	if update.ErrorMessage != nil {
		return "", errors.New(*update.ErrorMessage)
	} else if update.AckMessage != nil {
		return *update.AckMessage, nil
	}
	panic("Exec result neither error nor ack")
}
