import React from 'react';
import ReactDOM from 'react-dom';
import GraphiQL from 'graphiql';

function renderGraphiqlForURL({ endpointURL, domElement }) {
  function fetcher(graphQLParams) {
    return fetch(endpointURL, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
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
