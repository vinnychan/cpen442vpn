const {app, BrowserWindow, ipcMain, dialog} = require('electron')

const spawn = require('child_process').spawn;

const goMain = spawn('go', ['run', 'main.go']);

let isServer = null;

let debugMode = "y";

goMain.stdin.setEncoding('utf8');

goMain.stdout.on('data', (data) => {
  console.log(`stdout: ${data}`);
  if (isServer && isServer != null) {
    win.webContents.send('serverReply', {message: data});
  }
  if (!isServer && isServer != null) {
    win.webContents.send('clientReply', {message: data});
  }
});

goMain.stderr.on('data', (data) => {
  if (data.toString().indexOf('exit') != -1) {
    let code = data.toString().replace(/[^0-9]/g,'');
    console.log('EXIT: Application exited with code '+ code);
    app.quit();
  } else {
    console.log(`stderr: ${data}`);
  }
});

goMain.on('close', (code) => {
  console.log(`child process exited with code ${code}`);
});

ipcMain.on('debugToggle', function(event, data){
  // console.log(data.type + data.ip + data.port + data.secret);
  if (data.value == true) {
    debugMode = "y";
  } else {
    debugMode = "n";
  }

});

ipcMain.on('clientStart', function(event, data){
  // console.log(data.type + data.ip + data.port + data.secret);
  if (isServer == null) {
    if (data.debugMode == true) {
      debugMode = "y";
    } else {
      debugMode = "n";
    }
    goMain.stdin.write(debugMode + "\n");
    goMain.stdin.write(data.type.toString() + "\n");
    goMain.stdin.write(data.ip.toString() + "\n");
    goMain.stdin.write(data.port.toString() + "\n");
    goMain.stdin.write(data.secret.toString()+ "\n");
    let reply = "Connected to server at " + data.ip.toString() + " on port " +
      data.port.toString();
    event.sender.send('clientReply',
      {message: reply});
    isServer = false;
  } else {
    if (isServer) {
      dialog.showErrorBox("Connection Error",
        "Already started as a server. Please open another instance of the app"+
          " to select between client or server");
    }
    else {
      dialog.showErrorBox("Connection Error",
        "Already started as a client. Please open another instance of the app"+
          " to select between client or server");
    }

  }

});

ipcMain.on('serverStart', function(event, data){
  // console.log(data.type + data.port + data.secret);
  if (isServer == null){
    if (data.debugMode == true) {
      debugMode = "y";
    } else {
      debugMode = "n";
    }
    goMain.stdin.write(debugMode + "\n");
    goMain.stdin.write(data.type.toString() + "\n");
    goMain.stdin.write(data.port.toString() + "\n");
    goMain.stdin.write(data.secret.toString() + "\n");
    let reply = "Server started using port " +
      data.port.toString();
    event.sender.send('serverReply',
      {message: reply});
    isServer = true;
  } else {
    if (isServer) {
      dialog.showErrorBox("Connection Error",
        "Already started as a server. Please open another instance of the app"+
          " to select between client or server");
    }
    else {
      dialog.showErrorBox("Connection Error",
        "Already started as a client. Please open another instance of the app"+
          " to select between client or server");
    }
  }
});

ipcMain.on('sendAction', function(event, data){
  console.log(data);
  goMain.stdin.write(data.message.toString() + "\n");
});

ipcMain.on('continueAction', function(event, data){
  goMain.stdin.write("\n")
});


// Keep a global reference of the window object, if you don't, the window will
// be closed automatically when the JavaScript object is garbage collected.
let win

function createWindow () {
  // Create the browser window.
  win = new BrowserWindow({width: 800, height: 900})

  // and load the index.html of the app.
  win.loadURL(`file://${__dirname}/index.html`)

  // Open the DevTools.
  win.webContents.openDevTools()

  // Emitted when the window is closed.
  win.on('closed', () => {
    // Dereference the window object, usually you would store windows
    // in an array if your app supports multi windows, this is the time
    // when you should delete the corresponding element.
    win = null
  })
}

// This method will be called when Electron has finished
// initialization and is ready to create browser windows.
// Some APIs can only be used after this event occurs.
app.on('ready', createWindow)

// Quit when all windows are closed.
app.on('window-all-closed', () => {
  // On macOS it is common for applications and their menu bar
  // to stay active until the user quits explicitly with Cmd + Q
  if (process.platform !== 'darwin') {
    app.quit()
  }
})

app.on('activate', () => {
  // On macOS it's common to re-create a window in the app when the
  // dock icon is clicked and there are no other windows open.
  if (win === null) {
    createWindow()
  }
})

// In this file you can include the rest of your app's specific main process
// code. You can also put them in separate files and require them here.
