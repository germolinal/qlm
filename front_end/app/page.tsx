"use client";
// pages/index.js
import { useEffect, useState, useRef } from "react";

export default function Home() {
  const [message, setMessage] = useState("Waiting for webhook event...");
  const ws = useRef(null);

  useEffect(() => {
    ws.current = new WebSocket("ws://localhost:3001/api/hook"); // Updated URL

    ws.current.onopen = () => {
      console.log("WebSocket connection opened");
    };

    ws.current.onmessage = (event) => {
      let data = JSON.parse(event.data);

      if (data.response) {
        setMessage(data.response);
      } else if (data.message && data.message.content) {
        setMessage(data.message.content);
      } else {
        console.log("data", data, typeof data);
        console.log("response", data.response, typeof data.response);
        console.log("message", data.message, typeof data.message);
        console.log(
          "content",
          data.message.content,
          typeof data.message.content,
        );
        throw new Error(`found unexpected signature: ${JSON.stringify(data)}`);
      }
    };

    ws.current.onclose = () => {
      console.log("WebSocket connection closed");
    };

    ws.current.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    return () => {
      if (ws.current && ws.current.readyState === WebSocket.OPEN) {
        ws.current.close();
      }
    };
  }, []);

  return (
    <div>
      <h1>Real-time Webhook Example (Next.js Backend)</h1>
      <p>{message}</p>
    </div>
  );
}
