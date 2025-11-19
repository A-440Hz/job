
import { useState } from 'react';
import { registerUser, loginUser } from './api/user';
import './App.css';
import { useNavigate } from 'react-router-dom';

function LoginPage() {
	const [username, setUsername] = useState('');
	const [password, setPassword] = useState('');
	const [email, setEmail] = useState('');
	const [error, setError] = useState<string | null>(null);
	const [success, setSuccess] = useState<string | null>(null);
	const [isRegister, setIsRegister] = useState(false);
	const navigate = useNavigate();

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setError(null);
		setSuccess(null);
		try {
			if (isRegister) {
				await registerUser(username, password, email || undefined);
				setSuccess('Registration successful!');
				navigate('/');
			} else {
				await loginUser(username, password)
				setSuccess('Login successful!');
				navigate('/');
			}
		} catch (err: any) {
			setError(isRegister ? err.message : 'Invalid username or password');
		}
	};

	return (
		<div className="flex items-center justify-center min-h-screen ">
			<form
				className="bg-orange-200 border-2 border-blue-200 rounded-xl shadow-lg px-8 py-8 w-full max-w-md flex flex-col items-center"
				onSubmit={handleSubmit}
			>
				<h2 className="text-3xl font-bold text-neutral-700 mb-6 text-center select-none">
					{isRegister ? 'Register' : 'Login'}
				</h2>
				<div className="w-full mb-4">
					<label className="block text-left text-md font-semibold text-neutral-700 mb-1" htmlFor="username">Username</label>
					<input
						id="username"
						className="input-box px-3 py-2 rounded w-full border-2 border-blue-200 text-neutral-600 focus:border-blue-400 focus:outline-none"
						type="text"
						value={username}
						onChange={e => setUsername(e.target.value)}
						required
						autoComplete="username"
					/>
				</div>
				<div className="w-full mb-4">
					<label className="block text-left text-md font-semibold text-neutral-700 mb-1" htmlFor="password">Password</label>
					<input
						id="password"
						className="input-box px-3 py-2 rounded w-full border-2 border-blue-200 text-neutral-600 focus:border-blue-400 focus:outline-none"
						type="password"
						value={password}
						onChange={e => setPassword(e.target.value)}
						required
						autoComplete={isRegister ? 'new-password' : 'current-password'}
					/>
				</div>
				{isRegister && (
					<div className="w-full mb-4">
						<label className="block text-left text-md font-semibold text-neutral-700 mb-1" htmlFor="email">Email <span className="text-xs text-gray-400">(optional)</span></label>
						<input
							id="email"
							className="input-box px-3 py-2 rounded w-full border-2 border-blue-200 text-neutral-700 focus:border-blue-400 focus:outline-none"
							type="email"
							value={email}
							onChange={e => setEmail(e.target.value)}
							autoComplete="email"
						/>
					</div>
				)}
				{error && <div className="text-red-500 mb-2 text-center">{error}</div>}
				{success && <div className="text-green-600 mb-2 text-center">{success}</div>}
				<button
					type="submit"
					className="w-full py-2 mt-2 mb-2 rounded bg-slate-600 text-white font-semibold hover:bg-indigo-700 transition-colors"
				>
					{isRegister ? 'Register' : 'Login'}
				</button>
				<button
					type="button"
					className="w-full py-2 rounded border-2 border-indigo-400 text-indigo-700 font-semibold hover:bg-indigo-50 mt-1"
					onClick={() => setIsRegister(!isRegister)}
					title={isRegister ? 'Log in to an existing account' : 'Create a new account and secure your current progress'}
				>
					{isRegister ? 'Switch to Login page' : 'Register a new account'}
				</button>
			</form>
		</div>
	);
}

export default LoginPage;
