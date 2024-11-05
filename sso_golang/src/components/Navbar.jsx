import { Link, Outlet } from "react-router-dom";

const Navbar = () => {
	return (
		<>
			<nav className="bg-gray-900 text-white py-4">
				<div className="container mx-auto flex justify-between items-center">
					<Link to="/" className="text-xl font-bold">
						SSO Golang
					</Link>
					<div>
						<Link to="/signin" className="mr-4 hover:text-gray-300">
							Sign In
						</Link>
						<Link to="/signup" className="hover:text-gray-300">
							Sign Up
						</Link>
					</div>
				</div>
			</nav>
			<Outlet />
		</>
	);
};

export default Navbar;
