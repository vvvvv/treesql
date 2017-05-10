import { storeCommand } from './commandStorage';

export function sendCommand() {
  return (dispatch, getState) => {
    const command = getState().ui.command;
    dispatch(addMessage(command, 'client'));
    storeCommand(command);
    dispatch({
      type: 'UPDATE_COMMAND',
      newValue: ''
    });
    window.SOCKET.send(command);
  };
}

export function addMessage(message, source) {
  return function(dispatch) {
    var maybeJSON;
    try {
      maybeJSON = JSON.parse(message)
    } catch (e) {
      maybeJSON = message
    }
    dispatch({
      type: 'ADD_MESSAGE',
      message: maybeJSON,
      source
    });
    if (source === 'server') {
      console.log('scroll', document.body.scrollHeight);
      window.scrollTo(0, document.body.scrollHeight - window.innerHeight - 250);
    }
  }
}
