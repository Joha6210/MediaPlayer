const API_BASE = 'http://127.0.0.1:8090';

export async function getState() {
  const response = await fetch(`${API_BASE}/api/state`);
  if (!response.ok) {
    throw new Error(`State fetch failed: ${response.status}`);
  }
  return response.json();
}

export async function getSources(){
  const response = await fetch(`${API_BASE}/api/source/list`);
  if (!response.ok) {
    throw new Error(`Sources fetch failed: ${response.status}`);
  }
  return response.json();
}

export async function selectSource(payload: any) {
  const response = await fetch(`${API_BASE}/api/source/select`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  if (!response.ok) {
    throw new Error(await response.text());
  }
  return response.json();
}

export async function getStations() {
  const response = await fetch(`${API_BASE}/api/stations`, {
    method: 'GET',
    headers: { 'Content-Type': 'application/json' }
  });
  if (!response.ok) {
    throw new Error(`Stations fetch failed: ${response.status}`);
  }
  return response.json();
}

export async function setVolume(volume: number) {
  const response = await fetch(`${API_BASE}/api/player/volume`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ volume })
  });
  if (!response.ok) {
    throw new Error(await response.text());
  }
  return response.json();
}

export function openStateSocket(onState: (state: any) => void) {
  const ws = new WebSocket('ws://127.0.0.1:8090/ws');
  ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    onState(data);
  };
  return ws;
}
