const ipcRenderer = require('electron').ipcRenderer;

const clientForm = document.getElementById('client-config-form');

clientForm.addEventListener('submit', (evt)=>{
  evt.preventDefault();
  //TODO NULL CHECK
  let ip = clientForm.elements['ipaddress'].value;
  let port = clientForm.elements['port'].value;
  let secret = clientForm.elements['secret'].value;
  let debugMode = clientForm.elements['debug'][0].checked;
  //console.log( ip + port + secret );
  ipcRenderer.send('clientStart', 
    {type: 'client',
      debugMode: debugMode,
      ip: ip,
      port: port,
      secret: secret}
  );
}, false);


const serverForm = document.getElementById('server-config-form');
serverForm.addEventListener('submit', (evt)=>{
  //TODO NULL CHECK
  evt.preventDefault();
  let port = serverForm.elements['port'].value;
  let secret = serverForm.elements['secret'].value;
 let debugMode = serverForm.elements['debug'][0].checked;

  ipcRenderer.send('serverStart',
    {type: 'server',
      debugMode: debugMode,
      port: port,
      secret: secret}
  );
}, false);


const clientSendButton = document.getElementById('client-send-button');
//console.log(clientSendButton);
clientSendButton.addEventListener('click', ()=>{
  let message = document.getElementById('client-message').value;
  console.log(message);
  ipcRenderer.send('sendAction', {message:message});
});
const clientContinueButton = document.getElementById('client-continue-button');
//console.log(clientContinueButton);
clientContinueButton.addEventListener('click', ()=>{
  ipcRenderer.send('continueAction', '');
});

const serverSendButton = document.getElementById('server-send-button');
//console.log(serverSendButton);
serverSendButton.addEventListener('click', ()=>{
  let message = document.getElementById('server-message').value;
  console.log(message);
  ipcRenderer.send('sendAction', {message:message});
});
const serverContinueButton = document.getElementById('server-continue-button');
//console.log(serverContinueButton);
serverContinueButton.addEventListener('click', ()=>{
  ipcRenderer.send('continueAction', '');
});

ipcRenderer.on('clientReply', (event,data)=>{
  let clientTextarea = document.getElementById('client-textarea');
  let val = clientTextarea.value;
  clientTextarea.innerHTML= val + "\n" + data.message.toString();
  clientTextarea.scrollTop = clientTextarea.scrollHeight;
});

ipcRenderer.on('serverReply', (event,data)=>{
  let serverTextarea = document.getElementById('server-textarea');
  let val = serverTextarea.value;
  serverTextarea.innerHTML= val + "\n" + data.message.toString();
  serverTextarea.scrollTop = serverTextarea.scrollHeight;
});

