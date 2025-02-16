import { NextRequest, NextResponse } from "next/server";
import { WebSocketServer } from "ws";

class ConnectionHandler {
  private ws: WebSocket;

  constructor(ws: WebSocket) {
    this.ws = ws;

    this.ws.on("message", this.handleWebSocketMessage.bind(this));
    this.ws.on("close", this.handleWebSocketClose.bind(this));
    this.ws.on("error", this.handleWebSocketError.bind(this));
  }

  private handleWebSocketMessage(message: Buffer): void {
    console.log(
      `received '${message}'... but this is expected to be a one-way communication`,
    );
  }

  private handleWebSocketClose(): void {
    console.log("WebSocket connection closed...");
  }
  private handleWebSocketError(e: any) {
    console.error("Error on websocket", e);
  }

  send(message: any) {
    this.ws.send(JSON.stringify(message));
  }
}

const wss = new WebSocketServer({ port: 3001 });
let handler: ConnectionHandler | undefined = undefined;
wss.on("connection", (ws: WebSocket) => {
  handler = new ConnectionHandler(ws);
});

export async function POST(req: NextRequest) {
  if (handler) {
    let data = await req.json();
    handler.send(data);
    return NextResponse.json({ msg: "posting message" });
  } else {
    return NextResponse.json({ msg: "no handler" });
  }
}
