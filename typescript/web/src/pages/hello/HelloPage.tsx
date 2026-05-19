import React, { useState, useEffect } from 'react';

interface HelloResponse {
  result: {
    code: number;
    message: string;
  };
  message: string;
  timestamp: string;
}

export function HelloPage() {
  const [goMessage, setGoMessage] = useState<string>('');
  const [pyMessage, setPyMessage] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');

  useEffect(() => {
    const fetchHello = async () => {
      setLoading(true);
      setError('');

      try {
        const [goRes, pyRes] = await Promise.all([
          fetch('http://localhost:8080/hello'),
          fetch('http://localhost:8081/hello'),
        ]);

        if (goRes.ok) {
          const goData: HelloResponse = await goRes.json();
          setGoMessage(goData.message);
        }

        if (pyRes.ok) {
          const pyData: HelloResponse = await pyRes.json();
          setPyMessage(pyData.message);
        }
      } catch (err) {
        setError('无法连接到后端服务');
      } finally {
        setLoading(false);
      }
    };

    fetchHello();
  }, []);

  if (loading) {
    return <div data-testid="hello-loading">加载中...</div>;
  }

  if (error) {
    return <div data-testid="hello-error">{error}</div>;
  }

  return (
    <div data-testid="hello-page">
      <h1>Hello World</h1>
      <div>
        <h2>Golang Service</h2>
        <p data-testid="go-message">{goMessage || 'N/A'}</p>
      </div>
      <div>
        <h2>Python Service</h2>
        <p data-testid="py-message">{pyMessage || 'N/A'}</p>
      </div>
    </div>
  );
}
