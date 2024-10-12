class MySocket {
  constructor() {
    this.mysocket = null;
    this.vMsgContainer = document.getElementById("msgcontainer");
    this.vMsgIpt = document.getElementById("ipt");
  }

  showMessage(text, myself) {
    var div = document.createElement("div");
    div.innerHTML = text;
    var cself = myself ? "self" : "";
    div.className = "msg " + cself;
    this.vMsgContainer.appendChild(div);
  }

  send() {
    var txt = this.vMsgIpt.value;
    this.showMessage("<b>Me</b> " + txt, true);
    this.mysocket.send(txt);
    this.vMsgIpt.value = "";
  }

  keypress(e) {
    if (e.keyCode == 13) {
      this.send();
    }
  }

  connectSocket() {
    console.log("connecting socket");
    const host = WEBSOCKET_URL;
    var socket = new WebSocket(host + "/socket");
    this.mysocket = socket;

    socket.onmessage = (e) => {
      this.showMessage(e.data, false);
    };
    socket.onopen = () => {
      console.log("socket opened");
    };
    socket.onclose = () => {
      console.log("socket closed");
    };
  }
}
