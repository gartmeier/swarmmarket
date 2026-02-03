import { useState } from 'react';
import { Flag, X, Send, Loader2 } from 'lucide-react';

interface ReportButtonProps {
  itemType: 'listing' | 'request' | 'auction' | 'agent';
  itemId: string;
}

const reportReasons = [
  { value: 'spam', label: 'Spam or misleading' },
  { value: 'inappropriate', label: 'Inappropriate content' },
  { value: 'fraud', label: 'Suspected fraud' },
  { value: 'duplicate', label: 'Duplicate listing' },
  { value: 'other', label: 'Other' },
];

export function ReportButton({ itemType, itemId }: ReportButtonProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [selectedReason, setSelectedReason] = useState('');
  const [details, setDetails] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = async () => {
    if (!selectedReason) return;

    setSubmitting(true);
    // TODO: Implement actual report submission to API
    // For now, simulate API call
    await new Promise(resolve => setTimeout(resolve, 1000));

    console.log('Report submitted:', { itemType, itemId, reason: selectedReason, details });
    setSubmitting(false);
    setSubmitted(true);

    setTimeout(() => {
      setIsOpen(false);
      setSubmitted(false);
      setSelectedReason('');
      setDetails('');
    }, 2000);
  };

  return (
    <>
      <button
        onClick={() => setIsOpen(true)}
        className="flex items-center gap-2 px-4 h-10 rounded-lg transition-colors hover:bg-[#334155]"
        style={{ backgroundColor: 'transparent' }}
      >
        <Flag className="w-4 h-4 text-[#64748B]" />
        <span className="text-[#64748B] text-sm font-medium">Report</span>
      </button>

      {/* Modal Overlay */}
      {isOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center"
          style={{ backgroundColor: 'rgba(0, 0, 0, 0.7)' }}
          onClick={() => setIsOpen(false)}
        >
          {/* Modal Content */}
          <div
            className="w-full max-w-md mx-4"
            style={{
              backgroundColor: '#1E293B',
              borderRadius: '16px',
              border: '1px solid #334155',
            }}
            onClick={(e) => e.stopPropagation()}
          >
            {/* Header */}
            <div
              className="flex items-center justify-between px-6 py-4"
              style={{ borderBottom: '1px solid #334155' }}
            >
              <h3 className="text-lg font-semibold text-white">Report {itemType}</h3>
              <button
                onClick={() => setIsOpen(false)}
                className="w-8 h-8 rounded-lg flex items-center justify-center hover:bg-[#334155] transition-colors"
              >
                <X className="w-4 h-4 text-[#94A3B8]" />
              </button>
            </div>

            {/* Body */}
            <div className="p-6 flex flex-col gap-5">
              {submitted ? (
                <div className="flex flex-col items-center gap-3 py-4">
                  <div
                    className="w-12 h-12 rounded-full flex items-center justify-center"
                    style={{ backgroundColor: 'rgba(34, 197, 94, 0.2)' }}
                  >
                    <Flag className="w-6 h-6 text-[#22C55E]" />
                  </div>
                  <p className="text-white font-medium">Report submitted</p>
                  <p className="text-[#64748B] text-sm text-center">
                    Thank you for helping keep SwarmMarket safe. We'll review your report.
                  </p>
                </div>
              ) : (
                <>
                  <div className="flex flex-col gap-3">
                    <label className="text-sm font-medium text-[#94A3B8]">
                      Why are you reporting this {itemType}?
                    </label>
                    <div className="flex flex-col gap-2">
                      {reportReasons.map((reason) => (
                        <button
                          key={reason.value}
                          onClick={() => setSelectedReason(reason.value)}
                          className="flex items-center gap-3 p-3 rounded-lg transition-colors text-left"
                          style={{
                            backgroundColor: selectedReason === reason.value ? '#334155' : '#0F172A',
                            border: `1px solid ${selectedReason === reason.value ? '#22D3EE' : '#334155'}`,
                          }}
                        >
                          <div
                            className="w-4 h-4 rounded-full flex items-center justify-center"
                            style={{
                              border: `2px solid ${selectedReason === reason.value ? '#22D3EE' : '#64748B'}`,
                            }}
                          >
                            {selectedReason === reason.value && (
                              <div
                                className="w-2 h-2 rounded-full"
                                style={{ backgroundColor: '#22D3EE' }}
                              />
                            )}
                          </div>
                          <span
                            className="text-sm"
                            style={{ color: selectedReason === reason.value ? '#FFFFFF' : '#94A3B8' }}
                          >
                            {reason.label}
                          </span>
                        </button>
                      ))}
                    </div>
                  </div>

                  <div className="flex flex-col gap-2">
                    <label className="text-sm font-medium text-[#94A3B8]">
                      Additional details (optional)
                    </label>
                    <textarea
                      value={details}
                      onChange={(e) => setDetails(e.target.value)}
                      placeholder="Provide more context about the issue..."
                      rows={3}
                      className="w-full px-4 py-3 bg-[#0F172A] border border-[#334155] rounded-lg text-white text-sm placeholder-[#64748B] resize-none outline-none focus:border-[#22D3EE] transition-colors"
                    />
                  </div>

                  <button
                    onClick={handleSubmit}
                    disabled={!selectedReason || submitting}
                    className="w-full h-11 rounded-lg flex items-center justify-center gap-2 text-white font-medium transition-opacity disabled:opacity-50 disabled:cursor-not-allowed"
                    style={{
                      background: selectedReason
                        ? 'linear-gradient(90deg, #22D3EE 0%, #A855F7 100%)'
                        : '#334155',
                    }}
                  >
                    {submitting ? (
                      <>
                        <Loader2 className="w-4 h-4 animate-spin" />
                        Submitting...
                      </>
                    ) : (
                      <>
                        <Send className="w-4 h-4" />
                        Submit Report
                      </>
                    )}
                  </button>
                </>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
