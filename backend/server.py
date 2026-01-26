"""
Wrapper to start the Go backend from Python/Uvicorn infrastructure
This file is required because the Emergent platform expects a Python backend
"""
import os
import subprocess
import signal
import sys

# Store the Go process
go_process = None

def signal_handler(signum, frame):
    """Handle shutdown signals"""
    print(f"Received signal {signum}, shutting down Go backend...")
    if go_process:
        go_process.terminate()
        go_process.wait()
    sys.exit(0)

# Register signal handlers
signal.signal(signal.SIGTERM, signal_handler)
signal.signal(signal.SIGINT, signal_handler)

def main():
    """Start the Go backend and keep this process alive"""
    global go_process
    
    print("Starting Pulse Control Plane (Go Backend)...")
    go_binary = "/app/backend/pulse-control-plane"
    
    # Check if binary exists
    if not os.path.exists(go_binary):
        print(f"ERROR: Go binary not found at {go_binary}")
        print("Please build the Go backend first: cd /app/backend && go build -o pulse-control-plane .")
        sys.exit(1)
    
    # Start the Go process
    try:
        go_process = subprocess.Popen(
            [go_binary],
            cwd="/app/backend",
            env=os.environ.copy()
        )
        
        print(f"Go backend started with PID {go_process.pid}")
        
        # Wait for the Go process to complete
        go_process.wait()
        
        # If we reach here, the process exited
        print(f"Go backend exited with code {go_process.returncode}")
        sys.exit(go_process.returncode)
        
    except Exception as e:
        print(f"ERROR starting Go backend: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
