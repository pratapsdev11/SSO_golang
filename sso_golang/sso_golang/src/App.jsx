import router from "./routes/Router";
import Footer from "./components/Footer";
import { RouterProvider } from "react-router-dom";

const App = () => {
	return (
		<div className="flex flex-col min-h-screen">
			<div className="flex-grow">
				<RouterProvider router={router} />
			</div>
			<Footer />
		</div>
	);
};

export default App;
