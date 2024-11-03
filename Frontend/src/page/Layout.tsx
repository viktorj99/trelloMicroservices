import React from 'react';
import { Outlet } from 'react-router-dom';
import Navbar from '../components/Navbar/Navbar';

const Layout: React.FC = () => {
	return (
		<div>
			<Navbar />
			<div style={{ padding: '20px' }}>
				<Outlet />
			</div>
		</div>
	);
};

export default Layout;
