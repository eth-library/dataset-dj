package mailHandler

const (
	APILinkContent = `
	<h1>Welcome to the Data DJ</h1>

	<p>Thanks for joining the Data DJ!</p>
	
	<p>
	Below is a <em>single-use</em> link that returns an API Key.
	This API Key allows your application to authenticate with the Data DJ API.
	<br/>
	<br/>
	The Key can only be viewed once. It should be saved somewhere securely (e.g. password manager), not disclosed to users or client side code, and should not be hardcoded or checked into repositories.</p>
	</p>
	<p>
	   <a href="%v" target="_">%v</a> <br/>
	   (click or copy & paste into your browser)
	</p>
	
	In case of issues, please contact us at contact@librarylab.ethz.ch`

	DownloadLinkContent = `
	<h1>Your Download was completed</h1>

	<p>Thanks for using the Data DJ!</p>
	
	<p>
	Please use the link below to retrieve the requested files.
	</p>
	<p>
	   <a href="%v" target="_">%v</a> <br/>
	   (click or copy & paste into your browser)
	</p>
	
	In case of issues, please contact us at contact@librarylab.ethz.ch
	`
)
