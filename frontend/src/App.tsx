import { BrowserRouter, Routes, Route } from 'react-router-dom';
import JobAppTrackerPage from './JobAppTrackerPage';
import Topbar from './Topbar';

const About = () => <h2 className='justify-self-center text-4xl'>About Page</h2>;

function App() {


  return (
    <BrowserRouter>
      <Topbar />
      <Routes>
        <Route path="/" element={<JobAppTrackerPage />} />
        <Route path="/About" element={<About />} />
      </Routes>
    </BrowserRouter>
  );
}
export default App;