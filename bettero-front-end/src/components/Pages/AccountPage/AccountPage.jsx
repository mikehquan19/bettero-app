import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Link } from 'react-router-dom';
import AccountDetail from '../../AccountDetail/AccountDetail';
import { expenseappClient } from '../../../provider/api';
import handleError from '../../../provider/handleError';
import { useMediaQuery } from '@uidotdev/usehooks';

import './AccountPage.css';

// AccountPage component displaying the details of all each account 
export default function AccountPage({ navbarWidth, titleHeight }) {
  const { type } = useParams();
  const cardType = type.charAt(0).toUpperCase() + type.slice(1);
  const isMobileDevice = useMediaQuery("only screen and (max-width:500px)");

  // user and the list of accounts of this type 
  const [accountList, setAccountList] = useState([]);

  // fetch data about the list of accounts 
  useEffect(() => {
    expenseappClient.get("/accounts")
      .then((response) => {
        setAccountList(response.data.filter(account => account.account_type === cardType));
      })
      .catch((error) => handleError(error));
  }, []);

  return (
    <div style={{
      marginTop: `${titleHeight}px`,
      marginLeft: `${navbarWidth}px`,
      width: `calc(100% - ${navbarWidth}px)`,
    }}>
      <Link to="/" style={{
          fontSize: "18px",
          display: "block",
          width: isMobileDevice ? "calc(98% + 20px)" : "calc(85% + 20px)",
          margin: "0 auto",
          color: "blue",
        }}
        onMouseOver={(e) => { e.target.style.color = "skyblue"; }}
        onMouseOut={(e) => { e.target.style.color = "blue"; }}
      >Go Back</Link>
      <h2>Track the expense behavior of the user on each account</h2>
      <div className="debit-account-list">
        <h3 className="debit-account-list-title">{cardType} Accounts: </h3>
        {accountList.map(account =>
          <AccountDetail key={account.id}
            accountInfo={account}
            type={cardType}
            handleChangeAccList={(newAccountList) => setAccountList(newAccountList)} />
        )}
      </div>
    </div>
  );
}