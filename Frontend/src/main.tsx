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
import {isManager, isMember } from './utils/authHelpers';
import Verification from './page/Verification';

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
				path: '/project/:id',
				element: <SingleProject />,
			},
			{
				path: '/project/create',
				element: isManager() ? <CreateProject /> : <Navigate to="/" />,
			},
			{
				path: '/projects',
				element: (isManager() || isMember()) ? <ProjectList /> : <Navigate to="/login" />,
			},
			{
				path: '/verification',
				element: <Verification />,
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
