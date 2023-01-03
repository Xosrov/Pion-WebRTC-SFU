"use strict"
// from https://stackoverflow.com/questions/105034/how-do-i-create-a-guid-uuid
function uuidv4() {
    return ([1e7]+-1e3+-4e3+-8e3+-1e11).replace(/[018]/g, c =>
        (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
    );
}

window.ws = null;
const stream = new MediaStream();
window.pc = new RTCPeerConnection({
    iceServers: [{
        "urls": ["stun:stun.l.google.com:19302"],
    }]
});
window.pc.ontrack = function (event) {
    console.log(event.track.kind);
    stream.addTrack(event.track);
};
window.pc.onicecandidate = function(candidate) {
    if (candidate == null) {
        return;
    }
    window.ws.send(JSON.stringify({type: "icecandidate", payload: candidate}));
};

window.onload = function (){
    // create video element
    var Vel = document.createElement("video");
    Vel.srcObject = stream;
    Vel.autoplay = true;
    Vel.controls = true;
    document.getElementById('stream').appendChild(Vel);
    document.getElementById("uuid").innerText = uuidv4();
}

window.startWS = () => {
    let ssrc = document.getElementById("ssrc").value;
    if (ssrc == "") {
        console.log("ssrc not provided!")
        return
    }
    let uuid = document.getElementById("uuid").innerText;
    window.ws = new WebSocket(`${location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.hostname}:${window.location.port}/ws?sid=${ssrc}&uid=${uuid}`);
    window.ws.onopen = function (evt) {
        console.log("OPENED WS");
        // enable rtc button
        document.getElementById("rtcb").disabled = false;
        // send ping
        setInterval(function () {window.ws.send("ping");}, 5000);
    }
    window.ws.onclose = function (evt) {
        console.log("CLOSED WS");
        window.ws = null;
    }
    window.ws.onmessage = async function (evt) {
        try {
            let responseJson = JSON.parse(evt.data);
            console.log(responseJson);
            if (responseJson.type == "sdp") { // sdp message
                await window.pc.setRemoteDescription(responseJson.payload);
                console.log(responseJson.payload)
                console.log("Set offer");
                await window.pc.setLocalDescription();
                console.log("Set answer");
                console.log(window.pc.localDescription);
                window.ws.send(JSON.stringify({
                    type: "sdp",
                    payload: window.pc.localDescription,
                }));
            } else if (responseJson.type == "icecandidate") { // ice candidate message
                window.pc.addIceCandidate(responseJson.payload).then(() => {
                    console.log("Added new Ice candidate:");
                });
            } else if (responseJson.type == "pong") {
            } else {
                console.log("unknown type received: " + responseJson.type);
            }
        } catch (error) {
            return;
        }
    }
    window.ws.onerror = function (evt) {
        console.log("WS ERROR: " + evt.data);
    }
}

window.startRTC = () => {
    if (window.ws === null) {
        console.log("ws not created yet!")
        return
    }
    window.ws.send(JSON.stringify({
        type: "startrtc",
    }));
};
