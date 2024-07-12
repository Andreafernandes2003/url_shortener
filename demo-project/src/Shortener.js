import React, { useState, useEffect } from 'react';
import axios from 'axios';
import InputForm from './components/InputForm';
import ShortenedUrl from './components/ShortenedUrl';
import URLHistory from './components/URLHistory';

function Shortener() {
  const [shortUrl, setShortUrl] = useState('');
  const [copied, setCopied] = useState(false);
  const [urlHistory, setUrlHistory] = useState([]);

  useEffect(() => {
    const storedHistory = localStorage.getItem('urlHistory');
    if (storedHistory) {
      setUrlHistory(JSON.parse(storedHistory));
    }
  }, []);

  const handleClearHistory = () => {
    setUrlHistory([]);
    localStorage.removeItem('urlHistory');
  };

  const handleSubmit = async (url) => {
    const apiUrl = 'http://localhost:9091/shorten';

    const formattedUrl = /^(https?|ftp):\/\//i.test(url) ? url : `http://${url}`;

    try {
      const response = await axios.post(
        apiUrl,
        { original_url: formattedUrl },
        {
          headers: {
            'Content-Type': 'application/json',
          },
        }
      );

      console.log('API Response:', response.data);

      if (response.data && response.data.link) {
        const shortenedUrl = response.data.link;
        setShortUrl(shortenedUrl);

        const historyItem = {
          originalUrl: url,
          shortUrl: shortenedUrl,
          createdAt: new Date().toISOString(),
        };

        const updatedHistory = [historyItem, ...urlHistory];
        setUrlHistory(updatedHistory);
        localStorage.setItem('urlHistory', JSON.stringify(updatedHistory));
      } else {
        console.error('Invalid API response structure:', response.data);
      }
    } catch (error) {
      console.error('Error shortening URL:', error);
    }
  };

  const handleCopy = () => {
    setCopied(true);
    setTimeout(() => {
      setCopied(false);
    }, 3000);
  };

  return (
    <div className="shortener-container">
      <h1>URL Shortener</h1>

      <InputForm onSubmit={handleSubmit} />

      {shortUrl && <ShortenedUrl url={shortUrl} onCopy={handleCopy} />}

      {copied && <p className="copy-message">URL copied to clipboard!</p>}

      {urlHistory.length > 0 && <URLHistory history={urlHistory} onClearHistory={handleClearHistory} />}
    </div>
  );
}

export default Shortener;
