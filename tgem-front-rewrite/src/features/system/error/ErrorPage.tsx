import React from 'react';
import { Link, useNavigate } from 'react-router-dom';

const ErrorPage: React.FC = () => {
  const navigate = useNavigate();
  return (
    <div id='error-page' className='flex flex-col gap-8 justify-center items-center h-screen'>
      <h1 className='text-4xl font-bold'>404</h1>
      <p>Страница не найдена</p>
      <div className="flex space-x-3">
        <Link to={"/"} className="bg-gray-800 text-white text-2xl py-2 px-4 rounded cursor-pointer hover:bg-gray-900">На главную</Link>
        <button onClick={() => navigate(-1)} className="bg-gray-800 text-white text-2xl py-2 px-4 rounded cursor-pointer hover:bg-gray-900">Назад</button>
      </div>
    </div>
  );
};

export default ErrorPage;