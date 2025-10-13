import { BrowserRouter, Routes, Route } from 'react-router-dom';
import JobAppTrackerPage from './JobAppTrackerPage';
import Topbar from './Topbar';
import { ScreenSizeProvider } from './ScreenSizeProvider';
import UserPage from './UserPage';
import LoginPage from './LoginPage';
import LootboxPage from './Lootbox';
import { CollectablesPage } from './Collectables';

const About = () => <h2 className='justify-self-center text-4xl'>About Page</h2>;

function App() {


  return (
    <BrowserRouter>
    <ScreenSizeProvider>
      <Topbar />
      <Routes>
        <Route path="/" element={<JobAppTrackerPage />} />
        <Route path="/profile" element={<UserPage />} />
        <Route path="/About" element={<About />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/Lootbox" element={<LootboxPage />} />
        <Route path="/Collection" element={<CollectablesPage />} />
      </Routes>
    </ScreenSizeProvider>
    </BrowserRouter>
  );
}
export default App;