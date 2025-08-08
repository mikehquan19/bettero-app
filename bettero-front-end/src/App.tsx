import Navbar from '@components/Navbar/Navbar';
import MainPage from '@Pages/MainPage/MainPage';
import AccountPage from '@Pages/AccountPage/AccountPage';
import SummaryPage from '@Pages/SummaryPage/SummaryPage';
import BudgetPage from '@Pages/BudgetPage/BudgetPage';
import InvestmentPage from '@Pages/InvestmentPage/InvestmentPage';
import LoginPage from '@Pages/LoginPage/LoginPage';
import PrivateRoute from '@provider/PrivateRoute';
import LogoutRoute from '@provider/LogoutRoute';
import LogoutForm from '@Forms/DeleteForms/LogoutForm';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {faBars, faX, faHouse, faChartSimple, faMoneyBill, faArrowTrendUp } from '@fortawesome/free-solid-svg-icons'
import { CSSProperties, MouseEvent, useState } from 'react';
import { useMediaQuery } from '@uidotdev/usehooks';
import { BrowserRouter, Routes, Route, Link } from 'react-router-dom';
import './App.css';

function HeaderBar() {
  const [menuPresent, setMenuPresent] = useState(false);
  const [logoutFormPresent, setLogoutFormPresent] = useState(false);
  const isMobileDevice = useMediaQuery("only screen and (max-width: 750px)");

  const commonStyle: CSSProperties = !isMobileDevice ? 
    {
      textDecoration: "none",
      fontSize: "16px",
      color: "white",
      margin: "5px 0",
      padding: "0 5px",
      borderRadius: "10px",
    } : {
      borderTop: "1px solid black",
      textAlign: "center",
      fontSize: "16px",
      color: "white",
      textDecoration: "none",
    };

  function handleMouseOver(e: MouseEvent<HTMLElement>) {
    e.currentTarget.style.backgroundColor = "rgba(0, 0, 0, 0.25";
  }

  function handleMouseOut(e: MouseEvent<HTMLElement>) {
    e.currentTarget.style.backgroundColor = "darkcyan";
  }

  return (
    <>
      <div className="header-bar">
        <div className="title-area">
          <button>
            <Link to="/" 
              onMouseOver={handleMouseOver} onMouseOut={handleMouseOut}
              style={{
                padding: "10px 5px",
                borderRadius: "10px",
                textDecoration: "none",
                color: "white",
              } as CSSProperties}> Bettero App </Link>
            </button>
          {isMobileDevice && <button onClick={() => isMobileDevice ? setMenuPresent(!menuPresent) : null}>
            {menuPresent ?
              <FontAwesomeIcon icon={faX} style={{color: "white", fontSize: "24px"}} /> :
              <FontAwesomeIcon icon={faBars} style={{color: "white", fontSize: "24px"}} />
            }
          </button>}
        </div>
        {(menuPresent || !isMobileDevice) && (
          <div className="menu-wrapper">
            <Link to="/" 
              style={commonStyle} 
              onMouseOver={handleMouseOver} onMouseOut={handleMouseOut}
            >
              <FontAwesomeIcon style={{ marginRight: "5px" }} icon={faHouse} />Main
            </Link>
            <Link to="/summary"
              style={commonStyle} 
              onMouseOver={handleMouseOver} onMouseOut={handleMouseOut}
            >
              <FontAwesomeIcon style={{ marginRight: "5px" }} icon={faChartSimple} />Summary
            </Link>
            <Link to="/budget" 
              style={commonStyle} 
              onMouseOver={handleMouseOver} onMouseOut={handleMouseOut}
            >
              <FontAwesomeIcon style={{ marginRight: "5px" }} icon={faMoneyBill} />Budget
            </Link>
            <Link to="/investment" 
              style={commonStyle} 
              onMouseOver={handleMouseOver} onMouseOut={handleMouseOut}
            >
              <FontAwesomeIcon style={{ marginRight: "5px" }} icon={faArrowTrendUp} />Investment
            </Link>
            <div 
              style={commonStyle} 
              onMouseOver={handleMouseOver} onMouseOut={handleMouseOut}
              onClick={() => setLogoutFormPresent(!logoutFormPresent)}
            > 
              LOG OUT
            </div>
          </div>
        )}
      </div>
      {logoutFormPresent && <LogoutForm onHide={() => setLogoutFormPresent(!logoutFormPresent)} />}
    </>
  );
}


/**
 * The entire app which routes to the right page
 * @returns 
 */ 
export default function App() {
  const isMediumDevice = useMediaQuery("only screen and (max-width: 800px)");
  const navbarWidth = !isMediumDevice ? 150 : 0;
  const titleHeight = !isMediumDevice ? 0 : 70;

  return (
    <div>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={
            <PrivateRoute>
              {isMediumDevice ? (<HeaderBar />) : (<Navbar />)}
              <MainPage navbarWidth={navbarWidth} titleHeight={titleHeight} />
            </PrivateRoute>
          } />
          <Route path="/accounts/:type" element={
            <PrivateRoute>
              {isMediumDevice ? (<HeaderBar />) : (<Navbar />)}
              <AccountPage navbarWidth={navbarWidth} titleHeight={titleHeight} />
            </PrivateRoute>
          } />
          <Route path="/summary" element={
            <PrivateRoute>
              {isMediumDevice ? (<HeaderBar />) : (<Navbar />)}
              <SummaryPage navbarWidth={navbarWidth} titleHeight={titleHeight} />
            </PrivateRoute>
          } />
          <Route path="/budget" element={
            <PrivateRoute>
              {isMediumDevice ? (<HeaderBar />) : (<Navbar />)}
              <BudgetPage navbarWidth={navbarWidth} titleHeight={titleHeight} />
            </PrivateRoute>
          } />
          <Route path="/investment" element={
            <PrivateRoute>
              {isMediumDevice ? (<HeaderBar />) : (<Navbar />)}
              <InvestmentPage navbarWidth={navbarWidth} titleHeight={titleHeight} />
            </PrivateRoute>
          } />
          <Route path="/login" element={<LoginPage type="LOGIN" />} />
          <Route path="/register" element={<LoginPage type="SIGNUP" />} />
          <Route path="/logout" element={<LogoutRoute />} />
        </Routes>
      </BrowserRouter>
    </div>
  );
}