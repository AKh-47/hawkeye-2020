import React from 'react';
import { Switch, Redirect } from 'react-router';
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
import { PrivateRoute } from './PrivateRoute';

const AppRouter = ({ match }) => {
	return (
		<Provider loginStore={store} shopStore={ShopStore}>
			<Switch>
				<Route path={'/register'} component={RegisterPage}></Route>
				<Route path={'/login'} component={LoginPage}></Route>
				<Route path={'/forgot'} component={ForgotPass}></Route>
				<Route path={'/reset'} component={ResetPass}></Route>
				<Route path={'/admin/addquestion'} component={AddQuestion}></Route>
				<PrivateRoute path={'/regions'} component={LandingPage}></PrivateRoute>
				<PrivateRoute path={'/game/:id'} component={Game}></PrivateRoute>
				<PrivateRoute path={'/shop'} component={ShopPage}></PrivateRoute>

				<Route path={'/'} component={LoginPage}></Route>
			</Switch>
		</Provider>
	);
};

export { AppRouter };
