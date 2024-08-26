import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import TransactionTable from '../../TransactionTable/TransactionTable.jsx';
import DeleteAccountForm from '../../Forms/DeleteForms/DeleteAccountForm/DeleteAccountForm.jsx';
import AccountForm from '../../Forms/AddForms/AccountForm/AccountForm.jsx';
import ChangeChart from '../../Charts/ChangeChart.jsx';
import CompositionChart from '../../Charts/CompositionChart.jsx';
import './AccountPage.css';
import { getElementAtEvent } from 'react-chartjs-2';
import {Link} from 'react-router-dom';
import { expenseappClient } from '../../../provider/api.js';
import { useMediaQuery } from '@uidotdev/usehooks';

// reformat the date for value of the date field in the form 
const reformatDate = (argDate) => {
    if (argDate !== null) {
        const dateArray = argDate.split("/");
        const month = dateArray[0];
        const day = dateArray[1];
        const year = dateArray[2];
        return (year + "-" + month + "-" + day);
    }
    return null;
}


// function to log the error in the console depending on the type of error 
const handleError = (error) => {
    if (error.response) {
        // The request was made and the server responded with a status code
        // that falls out of the range of 2xx
        console.log(error.response.data);
        console.log(error.response.status);
        console.log(error.response.headers);
    } else if (error.request) {
        // The request was made but no response was received
        console.log(error.request);
    } else {
        // Something happened in setting up the request that triggered an Error
        console.log('Error', error.message);
    }
    console.log(error.config);
}


// displaying the detail of each account 
const AccountDetailCard = ({ accountInfo, type = "Debit", handleChangeAccList}) => {
    // info of the account 
    const [accountData, setAccountData] = useState({...accountInfo});
    // the expense change and composition data for the charts 
    const [expenseChange, setExpenseChange] = useState({}); 
    const [expenseComposition, setExpenseComposition] = useState({});

    // fetch data about the expense change and composition from API 
    useEffect(() => {
        expenseappClient.get(`/accounts/${accountInfo.id}/summary`)
            .then((response) => {
                // set the expense change and expense composition
                console.log("get account's summary request status: " + response.status);
                const summaryData = response.data; 
                setExpenseChange(summaryData.change_percentage); 
                setExpenseComposition(summaryData.composition_percentage); 
            })
            .catch((error) => handleError(error));
    }, [])

    // call the correct API request to fetch the transactions data 
    const fetchTransactionData = (transType = "latest", category = null) => {
        if (transType === "latest") {
            return expenseappClient.get(`accounts/${accountInfo.id}/transactions`); 
        } else {
            return expenseappClient.get(`accounts/${accountInfo.id}/transactions/category/${category}`);
        }
    }

    // state of the transaction table and the 2 charts of the account 
    const [showButtonContent, setShowButtonContent] = useState("Show details");
    const [transTablePresent, setTransTablePresent] = useState(false);
    const [transTableContent, setTransTableContent] = useState({
        type: null,
        title: null,
        transactionList: [],
    });

    // state of the form
    const [updateFormPresent, setUpdateFormPresent] = useState(false);
    const [deleteFormPresent, setDeleteFormPresent] = useState(false);

    // list of expense categories 
    const categoryList = ["Grocery", "Dining", "Shopping", "Bills", "Gas", "Others", "Income"];
    // position elements of the card depending on the card's type 
    let topSize = type === "Debit" ? "60px" : "10px";
    let isCreditAccount = type === "Debit" ? false : true;

    // handle category click (same function)
    const handleCategoryClick = (e, argChartRef) => {
        const selectedElement = getElementAtEvent(argChartRef.current, e)[0]
        // if the element is not undefined 
        if (selectedElement) {
            const newCategory = categoryList[selectedElement.index];
            console.log(newCategory);
            fetchTransactionData("category", newCategory)
                .then((response) => {
                    setTransTableContent({
                        type: "category",
                        title: newCategory,
                        transactionList: response.data,
                    });
                    // checking 
                    console.log("get categorical transactions request status:" + response.status);
                })
                .catch((error) => handleError(error)); 
        }
        else {
            console.log("clicked on no field");
        }
    }

    // handle the click to show the transaction table 
    const handleShowTableClick = () => {
        if (!transTablePresent) {
            setShowButtonContent("Hide details");
            setTransTablePresent(true);
            // reset the content of the transaction table when show it again 
            fetchTransactionData()
                .then((response) => {
                    setTransTableContent({
                        type: "latest",
                        title: "Latest",
                        // original list of total transactions
                        transactionList: response.data, 
                    });
                    // for checking 
                    console.log("get latest transactions request status:" + response.status);
                })
                .catch((error) => handleError(error)); 
        }
        else {
            setShowButtonContent("Show details");
            setTransTablePresent(false);
        }
    }

    // handle the click to search for the transactions (same function)
    const handleSearchClick = (searchTransName, tableType) => {
        if (searchTransName !== "") {
            const filteredTransList = transTableContent.transactionList.filter(transaction =>
                transaction.description === searchTransName
            );
            setTransTableContent(previousData => ({
                ...previousData,
                transactionList: filteredTransList,
            }));
        }
        else {
            // restore the original transaction list 
            if (tableType === "latest") {
                fetchTransactionData()
                    .then((response) => {
                        setTransTableContent((previousData) => ({
                            ...previousData,
                            transactionList: response.data,
                        }));
                        // checking 
                        console.log("get latest transactions request status:" + response.status);
                    })
                    .catch((error) => handleError(error));
            }
            else {
                const currentCategory = transTableContent.title;
                fetchTransactionData("category", currentCategory)
                .then((response) => {
                    setTransTableContent((previousData) => ({
                        ...previousData,
                        transactionList: response.data,
                    }));
                    // checking 
                    console.log("get categorical transactions request status:" + response.status);
                })
                .catch((error) => handleError(error)); 
            }
        }
    }

    return (
        <div className="account-detail-card">
            <div className="account-info" >
                <div className="institution-name">{accountData.institution}</div>
                <div className="account-name-num">{accountData.name}
                    <span className="account-num">
                        (****{accountData.account_number})
                    </span>
                </div>
                {type === "Credit" && (
                    <>
                        {/* the info strictly for the credit accounts */}
                        <div className="due-date">
                            <span style={{
                                fontWeight: "normal",
                                fontSize: "15px",
                            }}>Due date: </span> {accountData.due_date}
                        </div>
                        <div className="credit-limit">
                            <span style={{
                                fontSize: "15px",
                                fontWeight: "normal",
                            }}>Credit limit: </span> ${accountData.credit_limit}
                        </div>
                    </>
                )}
                <div className="balance" style={{
                    position: "relative",
                    top: topSize,
                }}>
                    <span style={{
                        fontSize: "15px",
                        fontWeight: "normal",
                    }}>Balance: </span> ${accountData.balance}
                </div>
            </div>
            <div className="button-list">
                {/* The button to show details */}
                <button className="show-detail-button"
                    onClick={handleShowTableClick}>{showButtonContent}</button>

                {/* The button to update the account */}
                <button className="update-info-button"
                    onClick={() => setUpdateFormPresent(true)}>Update account</button>

                {/* The button to delete the account */}
                <button className="delete-button"
                    onClick={() => setDeleteFormPresent(true)}>Delete account</button>
            </div>
            {transTablePresent && (
                <div style={{ marginTop: "20px", textAlign: "center"}}>
                    <div className="charts">
                        <div>
                            <ChangeChart changeObject={expenseChange} />
                        </div>
                        <div>
                            <CompositionChart 
                                compositionObject={expenseComposition} 
                                handleCategoryClick={handleCategoryClick} />
                        </div>
                    </div>
                    <p style={{
                        color: "darkcyan",
                        fontSize: "18px",
                        fontWeight: "bold",
                    }}>Click on the each category in the pie chart to show the account's list of transactions</p>
                    {/* if the table's type is category, shows the go back button */}
                    {transTableContent.type === "category" && (
                        <button className="go-back-button"
                            onClick={() => {
                                // reset the content of the transaction table when show it again 
                                fetchTransactionData()
                                    .then((response) => {
                                        setTransTableContent({
                                            type: "latest",
                                            title: "Latest",
                                            // original list of total transactions
                                            transactionList: response.data,
                                        });
                                        // for checking 
                                        console.log("get latest transactions request status:" + response.status); 
                                    })
                                    .catch((error) => handleError(error)); 
                            }}>Go back to latest list</button>
                    )}
                    <TransactionTable
                        listName={"Account's " + transTableContent.title + " Transactions"}
                        transactionList={transTableContent.transactionList}
                        tableType={transTableContent.type}
                        handleSearchClick={handleSearchClick}
                    />
                </div>
            )}
            {updateFormPresent && (
                <>
                    <h4 style={{ textAlign: "center" }}>Update the info you need</h4>
                    <AccountForm 
                        isCreditAccount={isCreditAccount} 
                        type="UPDATE"
                        handleHideForm={() => setUpdateFormPresent(false)}
                        handleChangeAccInfo={(newAccountInfo) => setAccountData({...newAccountInfo})}
                        currentData={{
                            ...accountData, 
                            due_date: reformatDate(accountInfo.due_date),
                        }} />
                </>
            )}
            {deleteFormPresent && (<DeleteAccountForm 
                accountId={accountData.id} 
                accountType={type}
                handleChangeAccList={handleChangeAccList}
                handleHideForm={() => setDeleteFormPresent(false)} 
            />)}
        </div>
    );
}


// AccountPage component displaying the details of all each account 
const AccountPage = ({ navbarWidth, titleHeight }) => {
    const { type } = useParams(); 
    const cardType = type.charAt(0).toUpperCase() + type.slice(1);
    const isMobileDevice = useMediaQuery("only screen and (max-width:500px)");

    // user and the list of accounts of this type 
    const [accountList, setAccountList] = useState([]); 

    // fetch data about the list of accounts 
    useEffect(() => {
        expenseappClient.get("/accounts")
            .then((response) => {
                console.log("get accounts request status: " + response.status); 
                setAccountList(response.data.filter(account => 
                    account.account_type === cardType
                ))
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
                onMouseOver={(e)=>{e.target.style.color = "skyblue";}}
                onMouseOut={(e) => {e.target.style.color = "blue";}}
            >Go Back</Link>
            <h2>Track the expense behavior of the user on each account</h2>
            <div className="debit-account-list">
                <h3>{cardType} Accounts: </h3>
                {accountList.map(account =>
                    <AccountDetailCard key={account.id} 
                    accountInfo={account}
                    type={cardType}  
                    handleChangeAccList={(newAccountList) => setAccountList(newAccountList)} />
                )}
            </div>
        </div>
    );
}

export default AccountPage;