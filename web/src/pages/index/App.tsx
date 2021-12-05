import React, { useState, useCallback, useEffect } from 'react'
import logo from '@/logo.svg'
import './App.css'
import * as oto from '../../service/oto.gen'

function App() {
  const [count, setCount] = useState(0)
  let client = new oto.Client()
  client.basepath = 'http://localhost:8080/api/oto/'
  let greeter = new oto.GreeterService(client)

  useEffect(() => {
    console.log("callback")
    let req = new oto.GreetRequest()
    req.name = 'joman'
    greeter.greet(req).then(ok => {
      console.log(ok)
    })
  }, [count])

  return (
    <div className="App">
      <header className="App-header">
        <img src={logo} className="App-logo" alt="logo" />
        <p>Hello Vite + React!</p>
        <p>
          <button onClick={() => setCount((count) => count + 1)}>
            count is: {count}
          </button>
        </p>
        <p>
          Edit <code>App.tsx</code> and save to test HMR updates.
        </p>
        <p>
          <a
            className="App-link"
            href="https://reactjs.org"
            target="_blank"
            rel="noopener noreferrer"
          >
            Learn React
          </a>
          {' | '}
          <a
            className="App-link"
            href="https://vitejs.dev/guide/features.html"
            target="_blank"
            rel="noopener noreferrer"
          >
            Vite Docs
          </a>
        </p>
      </header>
    </div>
  )
}

export default App
