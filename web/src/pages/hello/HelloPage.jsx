import React, { useState } from 'react';

function HelloPage() {
  const [name, setName] = useState('CaiRobot');
  const [goResponse, setGoResponse] = useState('');
  const [pyResponse, setPyResponse] = useState('');

  const fetchFromGo = async () => {
    try {
      const res = await fetch(`http://localhost:8080/hello?name=${encodeURIComponent(name)}`);
      const data = await res.json();
      setGoResponse(`Golang: ${data.message}`);
    } catch (err) {
      setGoResponse('Golang: 服务未启动');
    }
  };

  const fetchFromPy = async () => {
    try {
      const res = await fetch(`http://localhost:8081/hello?name=${encodeURIComponent(name)}`);
      const data = await res.json();
      setPyResponse(`Python: ${data.message}`);
    } catch (err) {
      setPyResponse('Python: 服务未启动');
    }
  };

  return (
    <div style={{ padding: '20px', maxWidth: '600px', margin: '0 auto' }}>
      <h1>HelloWorld 验收测试</h1>
      <div style={{ margin: '20px 0' }}>
        <label>
          姓名:
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            style={{ marginLeft: '10px', padding: '5px' }}
          />
        </label>
      </div>
      <div style={{ margin: '10px 0' }}>
        <button onClick={fetchFromGo} style={{ padding: '10px 20px', marginRight: '10px' }}>
          调用 Golang 服务
        </button>
        <button onClick={fetchFromPy} style={{ padding: '10px 20px' }}>
          调用 Python 服务
        </button>
      </div>
      <div style={{ marginTop: '20px' }}>
        <p>{goResponse}</p>
        <p>{pyResponse}</p>
      </div>
    </div>
  );
}

export default HelloPage;
