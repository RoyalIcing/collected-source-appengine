import './main.pcss'

import React from 'react';
import ReactDOM from 'react-dom';
import GraphiQL from 'graphiql';

function renderGraphiqlForURL({ domElement, endpointURL, headers }) {
  function fetcher(graphQLParams) {
    return fetch(endpointURL, {
      method: 'POST',
      headers: Object.assign({ 'Content-Type': 'application/json' }, headers),
      body: JSON.stringify(graphQLParams),
    }).then(response => response.json());
  }

  ReactDOM.render(<GraphiQL fetcher={fetcher} editorTheme="solarized light" />, domElement);
}

const tasks = {
  renderGraphiqlForURL
}

if (window.collectedTasks) {
  function runTask({ method, params }) {
    tasks[method](params)
  }

  window.collectedTasks.forEach(runTask);
  window.collectedTasks = {
    push: runTask
  };
}
