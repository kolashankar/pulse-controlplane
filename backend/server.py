"""
Proxy wrapper for Go backend
This FastAPI app proxies all requests to the Go backend running on port 8081
Required because Emergent platform infrastructure expects Python/FastAPI
"""
from fastapi import FastAPI, Request, Response
from fastapi.responses import StreamingResponse
import httpx
import os
import subprocess
import threading
import time
import atexit

# Go backend URL
GO_BACKEND_URL = "http://localhost:8081"
go_process = None

def start_go_backend():
    """Start the Go backend in a background thread"""
    global go_process
    
    go_binary = "/app/backend/pulse-control-plane"
    
    if not os.path.exists(go_binary):
        print(f"ERROR: Go binary not found at {go_binary}")
        return
    
    print("Starting Go backend...")
    go_process = subprocess.Popen(
        [go_binary],
        cwd="/app/backend",
        env=os.environ.copy(),
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE
    )
    print(f"Go backend started with PID {go_process.pid}")

def stop_go_backend():
    """Stop the Go backend"""
    global go_process
    if go_process:
        print("Stopping Go backend...")
        go_process.terminate()
        go_process.wait(timeout=10)

# Register cleanup
atexit.register(stop_go_backend)

# Start Go backend in background thread
thread = threading.Thread(target=start_go_backend, daemon=True)
thread.start()

# Wait for Go backend to be ready
time.sleep(2)

# Create FastAPI app
app = FastAPI(title="Pulse Control Plane Proxy")

@app.api_route("/{path:path}", methods=["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"])
async def proxy(request: Request, path: str):
    """Proxy all requests to Go backend"""
    
    # Build target URL
    url = f"{GO_BACKEND_URL}/{path}"
    if request.url.query:
        url += f"?{request.url.query}"
    
    # Get request body
    body = await request.body()
    
    # Forward request to Go backend
    async with httpx.AsyncClient() as client:
        try:
            response = await client.request(
                method=request.method,
                url=url,
                headers=dict(request.headers),
                content=body,
                timeout=30.0
            )
            
            return Response(
                content=response.content,
                status_code=response.status_code,
                headers=dict(response.headers)
            )
        except Exception as e:
            return Response(
                content=f"Error proxying to Go backend: {str(e)}",
                status_code=503
            )

@app.get("/")
async def root():
    """Root endpoint"""
    return {"message": "Pulse Control Plane - Proxy to Go Backend", "backend_url": GO_BACKEND_URL}
