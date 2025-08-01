import { BrowserRouter, Routes, Route } from 'react-router-dom';
import JobAppTrackerPage from './JobAppTrackerPage';
import Topbar from './Topbar';
import { ScreenSizeProvider } from './ScreenSizeProvider';

const About = () => <h2 className='justify-self-center text-4xl'>About Page</h2>;

function App() {


  return (
    <BrowserRouter>
    <ScreenSizeProvider>
      <Topbar />
      <Routes>
        <Route path="/" element={<JobAppTrackerPage />} />
        <Route path="/About" element={<About />} />
      </Routes>
    </ScreenSizeProvider>
    </BrowserRouter>
  );
}
export default App;