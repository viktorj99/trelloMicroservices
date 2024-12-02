import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import 'antd/dist/reset.css';

import { createBrowserRouter, Navigate, RouterProvider } from 'react-router-dom';
import Home from './page/Home';
import Login from './page/Login';
import Registration from './page/Registration';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import Layout from './page/Layout';
import CreateProject from './page/CreateProject';
import SingleProject from './page/SingleProject';
import ProjectList from './page/Projects';
import { isManager, isMember } from './utils/authHelpers';
import Verification from './page/Verification';
import ForgotPasswordPage from './page/ForgotPasswords';
import ChangePasswordPage from './page/ChangePassword';
import SendMagicLink from './page/SendMagicLink';
import MagicLogin from './page/MagicLogin';
import Profile from './page/Profile';

export const router = createBrowserRouter([
	{
		path: '/',
		element: <Layout />,
		children: [
			{
				path: '/',
				element: <Home />,
			},
			{
				path: '/login',
				element: <Login />,
			},
			{
				path: '/registration',
				element: <Registration />,
			},
			{
				path: '/profile',
				element: <Profile />,
			},
			{
				path: '/project/:id',
				element: isManager() || isMember() ? <SingleProject /> : <Navigate to='/login' />,
			},
			{
				path: '/project/create',
				element: isManager() ? <CreateProject /> : <Navigate to='/' />,
			},
			{
				path: '/projects',
				element: isManager() || isMember() ? <ProjectList /> : <Navigate to='/login' />,
			},
			{
				path: '/verification',
				element: <Verification />,
			},
			{
				path: '/forgot-password',
				element: <ForgotPasswordPage />,
			},
			{
				path: '/change-password',
				element: <ChangePasswordPage />,
			},
			{
				path: '/magic-link',
				element: <SendMagicLink />,
			},
			{
				path: '/magic-login',
				element: <MagicLogin />,
			},
		],
	},
]);

const queryClient = new QueryClient();

createRoot(document.getElementById('root')!).render(
	<StrictMode>
		<QueryClientProvider client={queryClient}>
			<RouterProvider router={router} />
		</QueryClientProvider>
	</StrictMode>
);
