// src/components/auth/LoginForm.jsx
import React from 'react';
import { useAuth } from '../../contexts/AuthContext';

const LoginForm = () => {
  const { login } = useAuth();

  const handleMicrosoftLogin = async (e) => {
    e.preventDefault();
    await login();
  };

  return (
    <div className="mt-8 space-y-6">
      <button
        onClick={handleMicrosoftLogin}
        className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
      >
        Sign in with Microsoft
      </button>
    </div>
  );
};

export default LoginForm;
