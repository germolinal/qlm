"use client";
// pages/index.js
import { useEffect, useState, useRef } from "react";

export default function Home() {
  const [sentMessages, setSentMessages] = useState<string[]>([]);
  const [responses, setResponses] = useState<string[]>([]);
  const ws = useRef(null);

  useEffect(() => {
    ws.current = new WebSocket("ws://localhost:3001/api/hook"); // Updated URL

    ws.current.onopen = () => {
      console.log("WebSocket connection opened");
    };

    ws.current.onmessage = (event: any) => {
      let data = JSON.parse(event.data);
      if (data.response) {
        setResponses((old) => [...old, data.response]);
      } else if (data.message && data.message.content) {
        setResponses((old) => [...old, data.message.content]);
      } else {
        console.error("data", data, typeof data);
        console.error("response", data.response, typeof data.response);
        console.error("message", data.message, typeof data.message);
        console.error(
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

    ws.current.onerror = (error: any) => {
      console.error("WebSocket error:", error);
    };

    return () => {
      if (ws.current && ws.current.readyState === WebSocket.OPEN) {
        ws.current.close();
      }
    };
  }, []);

  const submit = async (v: string) => {
    setSentMessages((old) => [...old, v]);
    let ret = await fetch("api/msg", {
      method: "POST",
      body: JSON.stringify({
        webhook: "http://localhost:3000/api/hook",
        prompt: v,

        format: {
          type: "object",
          properties: {
            inappropriate: {
              type: "boolean",
              description:
                "is this email inappropriate for a professional situation?",
            },
            contains_pii: {
              type: "boolean",
              description:
                "does the email contain Personally Identifiable Information or client data?",
            },
          },
          required: ["inappropriate", "contains_pii"],
        },
      }),
    });
    if (!ret.ok) {
      throw new Error(JSON.stringify(ret));
    }
    let msg = await ret.json();
    console.log(msg);
  };
  return (
    <div className="flex w-full">
      <div className="h-full flex-grow">
        <ul>
          {sentMessages.map((m: string, i: number) => {
            return <li key={i}>{m}</li>;
          })}
        </ul>
        <textarea
          onKeyUp={async (e) => {
            if (e.key === "Enter") {
              let v = e.target.value.trim();
              e.target.value = "";
              await submit(v);
            }
          }}
          className="w-full outline-none resize-none"
          name="newmsg"
          id="new-msg"
          placeholder="write a message"
        ></textarea>
      </div>
      <div className="min-w-[200px] max-w-[50vw]">
        <ul>
          {responses.map((r, i) => {
            return (
              <li key={i} className="p-2">
                {r}
              </li>
            );
          })}
        </ul>
      </div>
    </div>
  );
}
