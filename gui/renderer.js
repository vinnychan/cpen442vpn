const ipc = require('electron').ipcRenderer;
const clientButton = document.getElementById('client-button');
clientButton.addEventListener('click', function(){
      ipc.once('actionReply', function(response){
                processResponse(response);
            })
      ipc.send('invokeAction', 'someData');
});
