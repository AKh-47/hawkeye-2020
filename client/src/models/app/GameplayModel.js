import {decorate, observable, action} from 'mobx';
import { get, post } from '../../utils/api';
const hawkResponses = {
	Correct: `Hawk approves!`,
	Wrong: `Hawk disapproves.`,
	Close: `Hawk thinks you're close.`,
};

class GameplayModel {
	level= 0
	region=0
	question= ''
	message= ''
	currentAnswer= ''
	attempts= []
	hints= []
	stats= {atPar:0,trailing:0,leading:0}
	setCurrentAnswer(newValue) {
		this.currentAnswer = newValue;
	}

	getQuestion=(region)=> {
		if(this.region===0){
		this.region=region;
		}
		
	get(`/api/question/fetch/${this.region}`).then(this.getQuestionControl);
	}

	getQuestionControl=(res)=> {
		console.log(res)
		if (res.success) {
			if (res.data.question) {
				this.question = res.data.question;
				this.level = res.data.level;
				if (res.data.hints.length > 0) {
					this.hints.replace(res.data.hints.map(hint => hint.hint));
				} else {
					this.hints.replace(['No hints yet']);
				}
			}
		}
	}

	submit(region) {
		console.log(region)
		if (this.currentAnswer.length === 0) return;
		post(`/api/question/answer`, { answer: this.currentAnswer,region:parseInt(region) }).then(
			this.submitControl
		);
		this.attempts.unshift(this.currentAnswer);
		this.attempts.replace(this.attempts.slice(0, 10));
	}

	 submitControl=(res)=> {
		 console.log(hawkResponses[res.data.split(' ')[0]])
		if (res.success) {
			this.currentAnswer = '';
			this.message = hawkResponses[res.data.split(' ')[0]];
			setTimeout(this.clearMessage, 1000);
			if (res.data === 'Correct Answer') {
				setTimeout(this.getQuestion, 1000);
				
			}
		}
	}

	clearMessage() {
		this.message = '';
	}

	getTries() {
		get(`/api/gameplay/submissions`).then(this.getTriesControl);
	}

	getTriesControl(res) {
		if (res.success) {
			this.attempts.replace(res.data.map(sub => sub.answer));
		}
	}

	getStats() {
		get(`/api/gameplay/stats`).then(this.getStatsControl);
	}

     getStatsControl(res) {
		if (res.success) {
			this.stats.atPar = res.data.atPar;
			this.stats.leading = res.data.leading;
			this.stats.trailing = res.data.trailing;
		}
	}
}
decorate(GameplayModel,{
    level:observable,
	question:observable,
	message:observable,
	region:observable,
	currentAnswer:observable,
	attempts:observable,
	hints:observable,
	stats:observable,
	getQuestion:action,
	submit:action
})
const store=new GameplayModel()
export default store;