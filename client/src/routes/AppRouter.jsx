import React from 'react';
import { Switch } from 'react-router';
import { RegisterPage } from '../pages/app/RegisterPage/RegisterPage';
import LoginPage from '../pages/app/LoginPage/LoginPage';
import Game from '../pages/app/Game/Game';
import { Route } from 'react-router-dom';
import { Provider } from 'mobx-react';
import { LandingPage } from '../pages/app/LandingPage';
import store from '../models/app/LoginModel';
import ShopStore from '../models/app/ShopModel';
import { ShopPage } from '../pages/app/ShopPage/ShopPage';
import { ForgotPass } from '../pages/app/ForgotAndReset/ForgotPass';
import { ResetPass } from '../pages/app/ForgotAndReset/ResetPass';
import AddQuestion from '../pages/app/Admin/AddQuestion';
import AddHints from '../pages/app/Admin/AddHints';
import { PrivateRoute } from './PrivateRoute';
import { AdminRoute } from './AdminRoute';
import Leaderboard from '../pages/app/Admin/Leaderboard';

const AppRouter = () => {
	return (
		<Provider loginStore={store} shopStore={ShopStore}>
			<Switch>
				<Route path={'/register'} component={RegisterPage}></Route>
				<Route path={'/login'} component={LoginPage}></Route>
				<Route path={'/forgot'} component={ForgotPass}></Route>
				<Route path={'/reset'} component={ResetPass}></Route>
				<Route path={'/game/:id'} component={Game}></Route>
				<PrivateRoute path={'/shop'} component={ShopPage}></PrivateRoute>
				<PrivateRoute path={'/regions'} component={LandingPage}></PrivateRoute>
				<AdminRoute path={'/admin/addquestion'} component={AddQuestion} />
				<AdminRoute path={'/admin/addhints'} component={AddHints} />
				<AdminRoute path={'/admin/leaderboard'} component={Leaderboard} />
				<Route exact path={'/'} component={LoginPage}></Route>
			</Switch>
		</Provider>
	);
};

export { AppRouter };
