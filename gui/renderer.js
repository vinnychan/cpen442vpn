const ipcRenderer = require('electron').ipcRenderer;

const clientForm = document.getElementById('client-config-form');

clientForm.addEventListener('submit', (evt)=>{
  evt.preventDefault();
  //TODO NULL CHECK
  let ip = clientForm.elements['ipaddress'].value;
  let port = clientForm.elements['port'].value;
  let secret = clientForm.elements['secret'].value;
  console.log( ip + port + secret );
  ipcRenderer.send('clientStart', 
    {type: 'client',
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

  ipcRenderer.send('serverStart',
    {type: 'server',
      port: port,
      secret: secret}
    );
}, false);


const testButton = document.getElementById('test-button');
console.log(testButton);
testButton.addEventListener('click', ()=>{
  ipcRenderer.send('invokeAction', 'server');
});
