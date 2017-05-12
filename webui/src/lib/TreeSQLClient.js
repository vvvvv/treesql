import _ from 'lodash';

export const SCHEMA_QUERY = `
  many __tables__ {
    name,
    primary_key,
    columns: many __columns__ {
      name,
      type,
      references
    }
  }
`;

class EventEmitter {

  constructor() {
    this.listeners = {};
  }

  // TODO: auto-reconnect

  on(event, listener) {
    var listeners = this.listeners[event];
    if (!listeners) {
      listeners = [];
      this.listeners[event] = listeners;
    }
    listeners.push(listener);
  }

  off(event, listener) {
    _.remove(this.listeners[event], listener);
  }

  _dispatch(event, value) {
    this.listeners[event].forEach((listener) => {
      listener(value);
    });
  }

}

class Channel extends EventEmitter {

  constructor(client, statementID) {
    super();
    this.client = client;
    this.statementID = statementID;
  }

  _dispatchUpdate(message) {
    this._dispatch('update', message);
  }

}

export default class TreeSQLClient extends EventEmitter {

  constructor(url) {
    super();
    this.nextStatementId = 0;
    this.channels = {};
    this.websocket = new WebSocket(url);
    this.websocket.addEventListener('open', (evt) => {
      this._dispatch('open', evt);
    });
    this.websocket.addEventListener('close', (evt) => {
      this._dispatch('close', evt);
    });
    this.websocket.addEventListener('error', (evt) => {
      this._dispatch('error', evt);
    });
    this.websocket.addEventListener('message', (message) => {
      const parsedMessage = JSON.parse(message.data);
      this.channels[parsedMessage.StatementID]._dispatchUpdate(parsedMessage.Message);
    });
  }

  readyState() {
    return this.websocket.readyState;
  }

  sendStatement(query) {
    this.websocket.send(query);
    const channel = new Channel(this, this.nextStatementId);
    this.channels[this.nextStatementId] = channel;
    this.nextStatementId++;
    return channel;
  }

}
