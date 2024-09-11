#very simple image generator server based on flux.1
from http.server import BaseHTTPRequestHandler, HTTPServer
from urllib.parse import urlparse, parse_qs
import json
import os
import datetime
from io import BytesIO

import argparse


import torch
torch.cuda.empty_cache()
from diffusers import FluxPipeline

pipe = FluxPipeline.from_pretrained("black-forest-labs/FLUX.1-schnell", torch_dtype=torch.bfloat16)
pipe.enable_model_cpu_offload() #save some VRAM by offloading the model to CPU. Remove this if you have enough GPU power
pipe.enable_sequential_cpu_offload()


class MyHandler(BaseHTTPRequestHandler):
    def do_POST(self):

        content_len = int(self.headers.get('Content-Length'))
        post_body = self.rfile.read(content_len)

        seed=-1

        img = pipe(
        str(post_body),
        num_inference_steps=4, #use a larger number if you are using [dev]
        generator=torch.Generator("cpu").manual_seed(seed)
        ).images[0]


        self.send_response(200)
        self.send_header("Content-type", "image/png")
        self.send_header('Access-Control-Allow-Origin', '*')
        self.end_headers()

        img_io = BytesIO()
        img.save(img_io, 'PNG')
        img_io.seek(0)
            
        self.wfile.write(img_io.read())


if __name__ == '__main__':

    parser = argparse.ArgumentParser(description='Example language translator server')
    parser.add_argument('-p', '--port', type=int, default=8800, help='TCP port number')
    parser.add_argument('-m', '--model', type=str, default='black-forest-labs/FLUX.1-schnell', help='Model name')
    parser.add_argument('-s', '--maxlength', type=int, default=400, help="seed")
    args = parser.parse_args()


server_address = ('', args.port)
httpd = HTTPServer(server_address, MyHandler)
print('Server running model'+ args.model  +'on http://127.0.0.1:'+str(args.port))
httpd.serve_forever()
