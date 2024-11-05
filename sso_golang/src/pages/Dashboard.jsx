const Dashboard = () => {
	return (
		<div className="container mx-auto px-4 py-8">
			<h1 className="text-3xl font-bold mb-4">Dashboard</h1>
			<div className="bg-white shadow-md rounded px-8 pt-6 pb-8 mb-4">
				<h2 className="text-2xl font-bold mb-4">User Profile</h2>
				<div className="mb-4">
					<label
						className="block text-gray-700 font-bold mb-2"
						htmlFor="name"
					>
						Name
					</label>
					<p className="text-gray-700">John Doe</p>
				</div>
				<div className="mb-4">
					<label
						className="block text-gray-700 font-bold mb-2"
						htmlFor="email"
					>
						Email
					</label>
					<p className="text-gray-700">johndoe@example.com</p>
				</div>
			</div>
		</div>
	);
};

export default Dashboard;
