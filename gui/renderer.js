const ipcRenderer = require('electron').ipcRenderer;

const clientButton = document.getElementById('client-config-form');
console.log(clientButton);
clientButton.addEventListener('submit', (evt)=>{
  evt.preventDefault();
  console.log("clicked");
  ipcRenderer.send('invokeAction', {client:'client'});
}, true);
const serverButton = document.getElementById('server-config-form');
console.log(serverButton);
serverButton.addEventListener('submit', ()=>{
  evt.preventDefault();
  ipcRenderer.send('invokeAction', {server: 'server'});
}, true);
const testButton = document.getElementById('test-button');
console.log(testButton);
testButton.addEventListener('click', ()=>{
  ipcRenderer.send('invokeAction', 'server');
});
