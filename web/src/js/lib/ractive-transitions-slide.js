var DEFAULTS = {
	duration: 300,
	easing: 'easeInOut'
};

var PROPS = [
	'height',
	'borderTopWidth',
	'borderBottomWidth',
	'paddingTop',
	'paddingBottom',
	'marginTop',
	'marginBottom'
];

var COLLAPSED = {
	height: 0,
	borderTopWidth: 0,
	borderBottomWidth: 0,
	paddingTop: 0,
	paddingBottom: 0,
	marginTop: 0,
	marginBottom: 0
};

export default function slide ( t, params ) {
	var targetStyle;

	params = t.processParams( params, DEFAULTS );

	if ( t.isIntro ) {
		targetStyle = t.getStyle( PROPS );
		t.setStyle( COLLAPSED );
	} else {
		// make style explicit, so we're not transitioning to 'auto'
		t.setStyle( t.getStyle( PROPS ) );
		targetStyle = COLLAPSED;
	}

	t.setStyle( 'overflowY', 'hidden' );

	t.animateStyle( targetStyle, params ).then( t.complete );
}
