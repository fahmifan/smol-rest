import React, { useState, useCallback, useEffect } from 'react'
import logo from '@/logo.svg'
import './App.css'
import * as oto from '../../service/oto.gen'

function App() {
  const [count, setCount] = useState(0)
  let client = new oto.Client()
  let smoler = new oto.SmolService(client)

  useEffect(() => {
    console.log("callback")
    let req = new oto.AddTodoRequest()
    req.item = 'todo ' + new Date()
    smoler.addTodo(req).then(res => {
      console.log(res)
    })
  }, [count])

  const urlParams = new URLSearchParams(window.location.search);
  console.log("param", urlParams.get("param"))

  return (
    <div className="App">
      <h1>Dashboard</h1>
    </div>
  )
}

export default App
