import { NextRequest, NextResponse } from "next/server";

export const POST = async (req: NextRequest) => {
  let body = await req.json();
  let ret = await fetch("http://127.0.0.1:8080/async_generate", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  if (!ret.ok) {
    return NextResponse.error();
  }
  let data = await ret.text();

  return NextResponse.json(data, { status: 202 });
};
