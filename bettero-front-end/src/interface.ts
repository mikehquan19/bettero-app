
export interface FinancialInfo {
    totalBalance: number, 
    totalAmountDue: number, 
    totalIncome: number, 
    totalExpense: number
}

export interface Account {
    id: number, 
    user: number // user id
    accountNumber: number, 
    name: string, 
    institution: string, 
    accountType: 'Debit' | 'Credit', 
    balance: number, 
    creditLimit: number | null, 
    dueDate: Date | null 
}

export interface Transaction {
    id: number, 
    description: string, 
    category: string
    amount: number, 
    occurDate: Date, 
    accountName: string
}

export interface Bill {
    id: number, 
    user: number, 
    payAccount: string, 
    payAccountName: string,
    description: string, 
    category: string, 
    amount: number, 
    dueDate: Date
}

export interface OverdueBillMessage {
    id: number, 
    user: number, 
    billDescription: string, 
    billAmount: number, 
    billDueDate: Date, 
    appearDate: Date
}

export interface Stock {
    id: number, 
    user: number, 
    corporation: string, 
    name: string,
    symbol: string, 
    shares: number, 
    previousClose: number, 
    currentClose: number, 
    change: number,
    open: number, 
    low: number, 
    high: number, 
    volume: number, 
    lastUpdatedDate: Date
}


export interface BudgetPlan {
    income: number, 
    expensePortion: number, 
    composition: BudgetComposition, 
    progress: CategoryObject
}

export interface CategoryObject {
    gas: number | CategoryProgress, 
    dining: number | CategoryProgress, 
    others: number | CategoryProgress, 
    grocery: number | CategoryProgress, 
    housing: number | CategoryProgress, 
    medical: number | CategoryProgress, 
    shopping: number | CategoryProgress, 
    automobile: number | CategoryProgress, 
    subscription: number | CategoryProgress
}

export interface BudgetComposition {
    goal: CategoryObject, 
    actual: CategoryObject
}

export interface CategoryProgress {
    budget: number, 
    current: number, 
    percentage: number
}

export interface Interval {
    firstDate: string, 
    lastDate: string
}

export interface PageProps {
    navbarWidth: number, 
    titleHeight: number
}