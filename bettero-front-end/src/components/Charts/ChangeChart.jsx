import {
	Chart as ChartJS,
	CategoryScale,
	LinearScale,
	BarElement,
	Title,
	Tooltip,
	Legend,
} from 'chart.js';
import { Bar } from 'react-chartjs-2';

// register the tool to be able to to use it in the hand-shakeable way 
ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend);

// chart indicating the change percentage from the previous month 
export default function ChangeChart({ intervalType = "month", changeObject }) {

	// list of categories of expense for labels 
	const categories = Object.keys(changeObject);
	const changeData = Object.values(changeObject);

	// data of chart 
	const chartData = {
		labels: categories,
		datasets: [
			{
				label: `change from last ${intervalType} to this ${intervalType}`,
				data: changeData,
				backgroundColor: () => { // the color of the bar based on the value of the change
					const elementBackgroundColor = [];
					changeData.forEach((change) => {
						if (change > 0) {
							elementBackgroundColor.push("rgba(0, 0, 255, 0.5)");
						} else {
							elementBackgroundColor.push("rgba(255, 0, 0, 0.5)");
						}
					}); 
					return elementBackgroundColor;
				},
        
				// the color of the bar when hovered over
				hoverBackgroundColor: "rgb(0, 0, 0, 0.5)",
				hoverBorderColor: "white",
				hoverBorderWidth: 2,
			},
		],
	};

	// options of the chart 
	const chartOptions = {
		// this is to gain control of the width and height of the canvas
		maintainAspectRatio: false,
		responsive: true,
		// make the bar chart horizontal  
		indexAxis: 'y',
		scales: {
			y: {
				grid: {
					color: "skyblue",
				}
			}
		},
		plugins: {
			// set the title for the chart 
			title: {
				display: true,
				color: "black",
				text: 'Changes in catergories from previous ' + intervalType,
				font: {
					size: 20,
				}
			},
		},
	};

	// render the bar chart 
	return <Bar data={chartData} options={chartOptions} />;
}