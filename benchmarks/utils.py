from matplotlib import rc 
import matplotlib.pyplot as plt
import csv

def fig_to_file(fig, filename, ext):
    fig.savefig("graphs/%s.%s" % (filename, ext), format=ext, bbox_inches='tight')
    
def set_paper_rcs():
	fig_font = {'family':'sans-serif','sans-serif':['Helvetica'],
           'serif':['Helvetica'],'size':9}
	rc('font',**fig_font)
	rc('legend',fontsize=9, handletextpad=0.5)
	rc('text', usetex=True)
	rc('figure', figsize=(3.33,2.22))
	#  rc('figure.subplot', left=0.10, top=0.90, bottom=0.12, right=0.95)
	rc('axes', linewidth=0.5, color_cycle= ['#496ee2', '#8e053b', 'm', '#ef9708', 'g', 'c'])
	rc('lines', linewidth=1)

def draw_simple_plot(xlabel,ylabel,title,xdata,ydata):
    fig = plt.figure()
    axes = fig.add_axes([0.1, 0.1, 0.8, 0.8])

    axes.set_xlabel(xlabel)
    axes.set_ylabel(ylabel)
    axes.set_title(title)
    axes.plot(xdata,ydata,"rx")
    return axes, fig 
    
def draw_line_graph(xlabel,ylabel,title,xdata,ydata):
    fig = plt.figure()
    axes = fig.add_axes([0.1, 0.1, 0.8, 0.8])

    axes.set_xlabel(xlabel)
    axes.set_ylabel(ylabel)
    axes.set_title(title)
    axes.plot(xdata,ydata)
    return axes, fig
    
def draw_lines_graph(xlabel,ylabel,title,xdata,ydata,lines):
    fig = plt.figure()
    axes = fig.add_axes([0.1, 0.1, 0.8, 0.8])

    axes.set_xlabel(xlabel)
    axes.set_ylabel(ylabel)
    axes.set_title(title)

    for line in lines:
        axes.plot(xdata[line],ydata[line]) 

    axes.legend(lines,loc=1,frameon=True)
    return axes, fig

def draw_cdf(xlabel,title,data):
    fig = plt.figure()
    axes = fig.add_axes([0.1, 0.1, 0.8, 0.8])
    axes.set_xlabel(xlabel)
    axes.set_ylabel('Cumlative Propability')
    axes.set_title(title)

    axes.set_xlim([np.percentile(data,0),np.percentile(data,97)])
    axes.set_ylim([0,100])
    
    sorted_data = data
    sorted_data.sort()
    size=len(sorted_data)
    cdf_y = []
    for y in range (1,size+1):
        cdf_y.append(y*100.0/size)

    axes.plot(sorted_data, cdf_y)
    return axes, fig

    
def draw_cdfs(xlabel,title,data,lines):
    fig = plt.figure()
    axes = fig.add_axes([0.1, 0.1, 0.8, 0.8])
    axes.set_xlabel(xlabel)
    axes.set_ylabel('Cumlative Propability')
    axes.set_title(title)

    axes.set_xlim([0,5])
    axes.set_ylim([0,100])
    
    sorted_data = {}
    cdf_y = {}

    for line in lines:
        sorted_data[line] = data[line]
        sorted_data[line].sort()
        cdf_y[line]=[]
        size=len(sorted_data[line])
        for y in range (1,size+1):
            cdf_y[line].append(y*100.0/size)
            
        axes.plot(sorted_data[line], cdf_y[line])

    axes.legend(lines,loc=1,frameon=True)
    return axes, fig

def draw_histo(xlabel,title,data,bins):
    fig = plt.figure()
    axes = fig.add_axes([0.1, 0.1, 0.8, 0.8])

    axes.set_xlabel(xlabel)
    axes.set_ylabel('Probability')
    axes.set_title(title)

    n, bins, patches = axes.hist(data,bins,facecolor='green', normed=True, alpha=0.2)
    return axes, fig
    
def draw_boxplot(ylabel,title,data):
    fig = plt.figure()
    axes = fig.add_axes([0.1, 0.1, 0.8, 0.8])

    axes.set_ylabel(ylabel)
    axes.set_title(title)
    
    axes.boxplot(data,showfliers=False)
    return axes, fig
    
def draw_boxplots(xlabel,ylabel,title,data,lines):
    fig = plt.figure()
    axes = fig.add_axes([0.1, 0.1, 0.8, 0.8])

    axes.set_ylabel(ylabel)
    axes.set_xlabel(xlabel)
    axes.set_title(title)
    
    axes.boxplot([ v for v in data.values() ],labels=lines,showfliers=False)
    return axes, fig
    
def read_results_file(filename):
    results = {'latency':[],'reqs':[],'time':[]}
    start_time = 0.0
    start_set = False
    with open(filename, newline='') as csvfile:
        for row in csv.reader(csvfile):
            # latency in ms
            val = int(row[2])/1000000 
            results['latency'].append(val)
            results['reqs'].append(row[1])

            secs = float(row[0].rsplit(" ")[1].rsplit(":")[2])
            mins = float(row[0].rsplit(" ")[1].rsplit(":")[1])*60.0 

            # we assume all experiments run within an hour
            if not(start_set):
                start_time = secs+mins # time in secs
                start_set = True

            results['time'].append(((secs+mins) - start_time)*1000)
    return results
def latency_to_throughput(data):
    total = np.sum(data)
    measurements = len(data)
    return (1000 * measurements / total)