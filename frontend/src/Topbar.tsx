import { useEffect, useState } from 'react';
import { useCollectablesData } from './CollectablesDataContext';
import { logoutUser } from './api/user';
import { NavLink } from 'react-router-dom';
import { useLocation, useNavigate } from 'react-router-dom';
import { useScreenSize } from './ScreenSizeProvider';

function Glyph() {
  const [source, setSource] = useState("/sb1rb.png");
  const handleMouseAway: React.MouseEventHandler<HTMLImageElement> = (event) => {
    event.preventDefault();
    setSource("/sb2rb.png");
  };
  const handleMouseOver: React.MouseEventHandler<HTMLImageElement> = (event) => {
    event.preventDefault();
    setSource("/sb1rb.png");
  };
  const handleMouseClick = () => {
    window.location.href = "/";
  }
  return <img className="flex select-none min-w-12 max-w-24 mr-0 hover:cursor-pointer" id="glyph" 
    onMouseEnter={handleMouseAway} 
    onMouseLeave={handleMouseOver}
    onClick={handleMouseClick} 
    src={source}
    draggable="false"
  /> 
}

export function Navbar({ user }: { user: any }) {
  const isDesktop = useScreenSize();
  const navClass = "flex text-m/6 hover:opacity-60"
  return <nav className={`flex border-2 justify-between px-3 py-1 rounded-lg bg-slate-700 border-violet-200 ${isDesktop ? "space-x-12 mx-4" : "space-x-1"}`}>
    <NavLink to='/' className={({ isActive }) =>
        isActive ? navClass + "text-amber-300" : navClass
      }>
      Tracker
    </NavLink>
    <NavLink to='/Collection' className={({ isActive }) =>
        isActive ? navClass+ "text-amber-300" : navClass
      }>
      Collection
    </NavLink>
    <NavLink to='/Lootbox' className={({ isActive }) =>
        isActive ? navClass + "text-amber-300" : navClass
      }>
       Lootbox <p className={`text-orange-300 ${user.inventory?.NumLootboxes > 0 && "ml-1"}`}> {(user.inventory?.NumLootboxes > 0)? `(${user.inventory?.NumLootboxes})` : ''}</p>
    </NavLink>
    <NavLink to='/About' className={({ isActive }) =>
        isActive ? navClass + "text-amber-300" : navClass
      }>
      About
    </NavLink>
    <NavLink to='/profile' className={({ isActive }) =>
        isActive ? navClass + "text-amber-300" : navClass
      }>
      Profile
    </NavLink>
  </nav>;
} 

export function Login({ user, onLogout }: { user: any; onLogout?: () => void }) {
  if (!user) return <div>???</div>;
  if (user.Registered === true) {
    return (
      <button
        className="text-text-secondary pr-1.5 md:pr-3 hover:opacity-60"
        id="navbar_sign_in_button"
        onClick={() => onLogout && onLogout()}
      >
        Logout
      </button>
    );
  } else {
    return <a className="text-text-secondary pr-1.5 md:pr-3 hover:opacity-60" id="navbar_sign_in_button" href="/login">Login/Register</a>
  }
}

export default function Topbar() {
  // const isDesktop = useScreenSize();
  const {user, error, refreshData} = useCollectablesData();
  const location = useLocation();
  const navigate = useNavigate();

  const onLogout = async () => {
    try {
      await logoutUser();
    } catch (err) {
      console.error('Logout failed:', err);
    } finally {
      // refresh client state and navigate home
      try {
        await refreshData();
      } catch (e) {
        console.error('refreshData failed after logout', e);
      }
      navigate('/');
    }
  };

  // Refresh on mount and when the route path changes. This keeps the
  // navbar counters (e.g. lootbox count) reasonably up-to-date when the
  // user navigates between pages.
  useEffect(() => {
    refreshData();
    // only re-run when pathname changes
  }, [location.pathname]);

  if (error) return <div>Error loading backend: {error}</div>;
  if (!user) return <div>???</div>; 
    return (
    <div className='flex h-[62px] min-w-full items-center border justify-between select-none z50'>
      <span> <Glyph /> </span>
      {/* <span className="md:flex select-none font-semibold text-3xl ml-0"> {isDesktop && "haotianswebsite.com"} </span> */}
      <Navbar user={user}/>
      <Login user={user} onLogout={onLogout} />
    </div>);
}

