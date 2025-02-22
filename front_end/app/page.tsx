'use client'
// pages/index.js
import { useEffect, useState, useRef } from 'react'
type Flags = { inappropriate: boolean; contains_pii: boolean }

export default function Home () {
  const [sentMessages, setSentMessages] = useState<string[]>([])
  const [responses, setResponses] = useState<Flags[]>([])
  const ws = useRef<WebSocket | undefined>(undefined)

  useEffect(() => {
    ws.current = new WebSocket('ws://localhost:3001') 

    ws.current.onopen = () => {
      console.log('WebSocket connection opened')
    }

    ws.current.onmessage = (event: any) => {
      let data = JSON.parse(event.data)
      if (data.response) {
        setResponses(old => [...old, JSON.parse(data.response)])
      } else if (data.message && data.message.content) {
        setResponses(old => [...old, JSON.parse(data.message.content)])
      } else {
        console.error('data', data, typeof data)
        console.error('response', data.response, typeof data.response)
        console.error('message', data.message, typeof data.message)
        console.error(
          'content',
          data.message.content,
          typeof data.message.content
        )
        throw new Error(`found unexpected signature: ${JSON.stringify(data)}`)
      }
    }

    ws.current!.onclose = () => {
      console.log('WebSocket connection closed')
    }

    ws.current!.onerror = (error: any) => {
      console.error('WebSocket error:', error)
    }

    return () => {
      if (ws.current && ws.current.readyState === WebSocket.OPEN) {
        ws.current.close()
      }
    }
  }, [])

  const submit = async (v: string) => {
    if (v.trim().length===0){
      return
    }
    setSentMessages(old => [...old, v])
    let ret = await fetch('api/msg', {
      method: 'POST',
      body: JSON.stringify({
        webhook: 'http://localhost:3000/api/hook',
        prompt: v,

        format: {
          type: 'object',
          properties: {
            inappropriate: {
              type: 'boolean',
              description:
                'is this email inappropriate for a professional situation?'
            },
            contains_pii: {
              type: 'boolean',
              description:
                'does the email contain Personally Identifiable Information or client data?'
            }
          },
          required: ['inappropriate', 'contains_pii']
        }
      })
    })
    if (!ret.ok) {
      throw new Error(JSON.stringify(ret))
    }
    let msg = await ret.json()
    console.log(msg)
  }
  return (
    <div className='p-2'>
      <textarea
        onKeyUp={async e => {
          if (e.key === 'Enter') {
            let v = e.target.value.trim()
            e.target.value = ''
            await submit(v)
          }
        }}
        className='w-full outline-none resize-none p-1 border border-gray rounded-md'
        name='newmsg'
        id='new-msg'
        placeholder='write a message'
      ></textarea>
      <div className='flex w-full'>
        <div className='h-full flex-grow'>
          <ul>
            {sentMessages.map((m: string, i: number) => {
              return (
                <li
                  className='bg-gray-100 py-1 px-2 my-1 rounded-md border border-gray-300'
                  key={i}
                >
                  {m}
                </li>
              )
            })}
          </ul>
        </div>
        <div className='min-w-[200px] max-w-[50vw]'>
          <ul>
            {responses.map((r, i) => {
              return (
                <li key={i} className='flex py-1 space-x-1 px-2 min-h-[1em]'>
                  {r.contains_pii && (
                    <span className='bg-red-500 text-white px-2 min-w-[2em] h-[2em] rounded-full'>
                      PII
                    </span>
                  )}
                  {r.inappropriate && (
                    <span className='bg-pink-950 text-white px-2 min-w-[2em] h-[2em] rounded-full'>
                      Inappropriate
                    </span>
                  )}
                </li>
              )
            })}
          </ul>
        </div>
      </div>
    </div>
  )
}
