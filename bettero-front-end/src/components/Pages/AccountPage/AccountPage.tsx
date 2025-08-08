import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Link } from 'react-router-dom';
import AccountDetail from '@components/AccountDetail/AccountDetail';
import { getUserAccounts } from '@provider/api';
import handleError from '@provider/handleError';
import { useMediaQuery } from '@uidotdev/usehooks';
import { Account } from '@interface';
import './AccountPage.scss';

interface AccountPageProps {
  navbarWidth: number, 
  titleHeight: number
}

/**
 * AccountPage component displaying the details of all each account 
 * @param {AccountPageProps} 
 * @returns {JSX.Element}
 */
export default function AccountPage({ navbarWidth, titleHeight }: AccountPageProps): JSX.Element {
  const { type } = useParams();
  const cardType = type!.charAt(0).toUpperCase() + type!.slice(1);
  const isMobileDevice = useMediaQuery("only screen and (max-width:500px)");

  // User and the list of accounts of this type 
  const [accountList, setAccountList] = useState<Account[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true); 

  // fetch data about the list of accounts 
  useEffect(() => {
    getUserAccounts()
      .then((response) => {
        const accountData = response.data as Account[];
        setAccountList(accountData.filter(account => account.accountType === cardType));
        setIsLoading(false);
      })
      .catch((error) => handleError(error));
  }, []);

  if (isLoading) return <div>...Loading</div>;
  return (
    <div 
      style={{ marginTop: `${titleHeight}px`, marginLeft: `${navbarWidth}px`, width: `calc(100% - ${navbarWidth}px)`}}
    >
      <Link to="/" className='link' 
        style={{ width: isMobileDevice ? "calc(98% + 20px)" : "calc(85% + 20px)" }}
        onMouseOver={(e) => e.currentTarget.style.color = "skyblue"}
        onMouseOut={(e) => e.currentTarget.style.color = "blue" }
      >
        Go Back
      </Link>

      <h2>Track the expense behavior of the user on each account</h2>

      <div className="debit-account-list">
        <h3 className="debit-account-list-title">{cardType} Accounts:</h3>
        {accountList.map(account =>
          <AccountDetail 
            key={account.id} type={cardType} accountInfo={account}
            onChangeAccList={(newAccountList: Account[]) => setAccountList(newAccountList)} />
        )}
      </div>
    </div>
  );
}