import _ from 'lodash';
import React, { Component } from 'react';
import { connect } from 'react-redux';
import classNames from 'classnames';
import Autocomplete from 'react-autocomplete';
import { sendStatementFromInput } from '../actions.js';
import StatementLog from './StatementLog';
import './REPL.css';

const WEBSOCKET_STATES = {
  [WebSocket.CLOSED]: 'CLOSED',
  [WebSocket.CONNECTING]: 'CONNECTING',
  [WebSocket.OPEN]: 'OPEN',
  [WebSocket.CLOSING]: 'CLOSING'
}

class App extends Component {
  
  componentDidMount() {
    document.getElementById('statement-input').focus(); // I forget how to do refs
  }

  render() {
    return (
      <div className="App">
        <div>
          Websocket state: {WEBSOCKET_STATES[this.props.ui.websocketState]}
        </div>
        <ol className="statements">
          {this.props.statements.map((statement) => (
            <li key={statement.id}>
              <div className="statement">{statement.statement}</div>
              <StatementLog updates={statement.updates} />
            </li>
          ))}
        </ol>
        <form onSubmit={this.props.onSubmit}>
          <div id="statement-input-container">
            <Autocomplete
              value={this.props.ui.statement}
              onChange={(evt, value) => this.props.updateStatement(value)}
              onSelect={(value) => this.props.updateStatement(value)}
              items={_.reverse(this.props.ui.statementHistory)}
              shouldItemRender={(item, value) => (item.indexOf(value) !== -1)}
              renderItem={(item, isHighlighted) => (
                <div className={classNames('statement-choice', { selected: isHighlighted })}>
                  {item}
                </div>)}
              getItemValue={_.identity}
              inputProps={{ size: 100, id: 'statement-input' }} />
          </div>
          <button
            disabled={this.props.ui.websocketState !== WebSocket.OPEN}>
            Send
          </button>
        </form>
      </div>
    );
  }
}

function mapStateToProps(state) {
  return state.state;
}

function mapDispatchToProps(dispatch) {
  return {
    onSubmit: (evt) => {
      evt.preventDefault();
      dispatch(sendStatementFromInput());
    },
    updateStatement: (newValue) => {
      dispatch({
        type: 'UPDATE_STATEMENT',
        newValue
      })
    }
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(App);
