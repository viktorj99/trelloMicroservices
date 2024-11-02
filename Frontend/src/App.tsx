import { useEffect, useState } from 'react';

interface JsonDataPlaceholder {
	userId: number;
	id: number;
	completed: boolean;
	title: string;
}

function App() {
	const [gas, setGas] = useState(0);
	const [jsonData, setJsonData] = useState<JsonDataPlaceholder | null>(null);

	const handleClick = () => {
		setGas((previousValue) => previousValue + 1);
		setGas((previousValue) => previousValue + 1);
	};

	useEffect(() => {
		const fetchData = async () => {
			try {
				const response = await fetch('https://jsonplaceholder.typicode.com/todos/1');
				const json = await response.json();
				setJsonData(json);
				console.log('aaaaaa');
			} catch (error) {
				console.error('Error fetching data:', error);
			}
		};

		fetchData();
	}, []);

	return (
		<>
			<div>{gas}</div>
			<button onClick={handleClick}>Click</button>
			<div>
				<p>{jsonData?.userId}</p>
				<p>{jsonData?.title}</p>
			</div>
		</>
	);
}

export default App;
