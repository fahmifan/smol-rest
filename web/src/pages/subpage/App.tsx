import React, { useState, useEffect } from 'react'
import logo from '@/logo.svg'
import './App.css'
import * as oto from '../../service/oto.gen'

function App() {
  const [count, setCount] = useState(0)
  let greeter = new oto.GreeterService(new oto.Client())
  useEffect(() => {
    let req = new oto.GreetRequest()
    req.name = 'joman'
    greeter.greet(req)
  }, [count])

  return (
    <div className="App">
      <h1>Subpages</h1>
    </div>
  )
}

export default App
