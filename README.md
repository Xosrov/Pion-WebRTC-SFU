# Pion-WebRTC-SFU  
All-in-one WebRTC SFU server, an overpowered version of [this example](https://github.com/pion/webrtc/tree/master/examples/rtp-to-webrtc).  

## Overview
Low-latency streaming of a singular audio/video pipeline.
1. ### Fast
    The underlying WebRTC protocol provides realtime media streaming capabilities.
2. ### Flexible
    Support for popular protocols like VPX, AV1 and H264/H265. Handles concurrent streaming from multiple servers to multiple users.
3. ### Expandable 
    Can be improved to your specific needs because of its modular structure.

## Usage
### 1. Configure .env and start server
`cp .env_sample .env`  
now edit the .env file with desired configuration. Note that RTP codec must match RTP stream contents. Start the server with `go run .` in project root.  

### 2. Start RTP stream  
Start an RTP stream from local or remote device and send the udp packets with mtu=1200 (check .env) to the server. GStreamer or FFmpeg can be used for this. Two simple examples runnning on the local device are shown below:

**GStreamer (VP8 - video only)**  
`gst-launch-1.0 videotestsrc ! video/x-raw,width=640,height=480,format=I420 ! vp8enc error-resilient=default keyframe-max-dist=10 auto-alt-ref=true cpu-used=8 deadline=1 ! rtpvp8pay mtu=1200 ssrc=12345 ! udpsink host=127.0.0.1 port=5004`

**FFmpeg (x264 - video only)**
`ffmpeg -re -f lavfi -i testsrc=size=640x480:rate=30 -pix_fmt yuv420p -c:v libx264 -g 10 -preset ultrafast -tune zerolatency -ssrc 12345 -f rtp 'rtp://127.0.0.1:5004?pkt_size=1200'`

### 3. Go to `localhost:8080`  
Enter the ssrc of the RTP stream and start the stream. In  the above RTP streams, ssrc is set to 12345; so enter that.
ssrc for video and audio tracks must be the same if both are used simultaneously. 


## TODO
- Better documentation
- More and better examples
- PLI handling
- Cleaner code
- More customization
- Better logging
- ...