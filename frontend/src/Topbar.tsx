import { useState } from 'react';
import { NavLink } from 'react-router-dom';

function Glyph() {
  const [source, setSource] = useState("/src/assets/sb1rb.png");
  const handleMouseAway: React.MouseEventHandler<HTMLImageElement> = (event) => {
    event.preventDefault();
    setSource("/src/assets/sb2rb.png");
  };
  const handleMouseOver: React.MouseEventHandler<HTMLImageElement> = (event) => {
    event.preventDefault();
    setSource("/src/assets/sb1rb.png");
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

export function Navbar() {
  const navClass = "m-1 text-m/6 hover:opacity-60 "
  return <nav className='border-2 justify-self-end pl-3 pr-3'>
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
      Open Lootbox
    </NavLink>
    <NavLink to='/About' className={({ isActive }) =>
        isActive ? navClass + "text-amber-300" : navClass
      }>
      About
    </NavLink>
  </nav>;
} 

export function Login() {
  return <a className="text-text-secondary pr-3 hover:opacity-60" id="navbar_sign_in_button" href="/login">Log in</a>
}

export default function Topbar() {
    return (
    <div className='flex h-[62px] items-center border justify-between select-none z50'>
      <span> <Glyph /> </span>
      <span className="hidden md:flex border select-none text-4xl ml-0"> Hi this is a topbar </span>
      <Navbar />
      <Login />
    </div>);
}

